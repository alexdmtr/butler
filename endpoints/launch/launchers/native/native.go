package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/itchio/httpkit/neterr"

	"github.com/itchio/pelican"

	"github.com/itchio/dash"

	"github.com/itchio/butler/butlerd/messages"
	"github.com/itchio/butler/filtering"

	"github.com/itchio/butler/butlerd"
	"github.com/itchio/butler/cmd/wipe"
	"github.com/itchio/butler/endpoints/launch"
	"github.com/itchio/butler/runner"
	"github.com/pkg/errors"
)

func Register() {
	launch.RegisterLauncher(launch.LaunchStrategyNative, &Launcher{})
}

type Launcher struct{}

var _ launch.Launcher = (*Launcher)(nil)

func (l *Launcher) Do(params *launch.LauncherParams) error {
	consumer := params.RequestContext.Consumer
	installFolder := params.InstallFolder

	cwd := installFolder
	_, err := filepath.Rel(installFolder, params.FullTargetPath)
	if err == nil {
		// if it's relative, set the cwd to the folder the
		// target is in
		cwd = filepath.Dir(params.FullTargetPath)
	}

	_, err = os.Stat(params.FullTargetPath)
	if err != nil {
		return errors.WithStack(err)
	}

	err = configureTargetIfNeeded(params)
	if err != nil {
		consumer.Warnf("Could not configure launch target: %s", err.Error())
	}

	err = fillPeInfoIfNeeded(params)
	if err != nil {
		consumer.Warnf("Could not determine PE info: %s", err.Error())
	}

	err = handlePrereqs(params)
	if err != nil {
		if be, ok := butlerd.AsButlerdError(err); ok {
			switch butlerd.Code(be.RpcErrorCode()) {
			case butlerd.CodeOperationAborted, butlerd.CodeOperationCancelled:
				return be
			}
		}

		consumer.Warnf("While handling prereqs: %+v", err)

		if neterr.IsNetworkError(err) {
			err = butlerd.CodeNetworkDisconnected
		}

		r, err := messages.PrereqsFailed.Call(params.RequestContext, &butlerd.PrereqsFailedParams{
			Error:      err.Error(),
			ErrorStack: fmt.Sprintf("%+v", err),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		if r.Continue {
			// continue!
			consumer.Warnf("Continuing after prereqs failure because user told us to")
		} else {
			// abort
			consumer.Warnf("Giving up after prereqs failure because user asked us to")
			return errors.WithStack(butlerd.CodeOperationAborted)
		}
	}

	envMap := make(map[string]string)
	for k, v := range params.Env {
		envMap[k] = v
	}

	// give the app its own temporary directory
	tempDir := filepath.Join(params.InstallFolder, ".itch", "temp")
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		consumer.Warnf("Could not make temporary directory: %s", err.Error())
	} else {
		defer wipe.Do(consumer, tempDir)
		envMap["TMP"] = tempDir
		envMap["TEMP"] = tempDir
		consumer.Infof("Giving app temp dir (%s)", tempDir)
	}

	var envKeys []string
	for k := range envMap {
		envKeys = append(envKeys, k)
	}
	consumer.Infof("Environment variables passed: %s", strings.Join(envKeys, ", "))

	// TODO: sanitize environment somewhat?
	envBlock := os.Environ()
	for k, v := range envMap {
		envBlock = append(envBlock, fmt.Sprintf("%s=%s", k, v))
	}

	const maxLines = 40
	stdout := newOutputCollector(maxLines)
	stderr := newOutputCollector(maxLines)

	fullTargetPath := params.FullTargetPath
	name := params.FullTargetPath
	args := params.Args

	if params.Candidate != nil && params.Candidate.Flavor == dash.FlavorLove {
		// TODO: add prereqs when that happens
		args = append([]string{name}, args...)
		name = "love"
		fullTargetPath = "love"
	}

	runParams := &runner.RunnerParams{
		RequestContext: params.RequestContext,
		Ctx:            params.Ctx,

		Sandbox: params.Sandbox,

		FullTargetPath: fullTargetPath,

		Name:   name,
		Dir:    cwd,
		Args:   args,
		Env:    envBlock,
		Stdout: stdout,
		Stderr: stderr,

		PrereqsDir:    params.PrereqsDir,
		Credentials:   params.Credentials,
		InstallFolder: params.InstallFolder,
		Runtime:       params.Runtime,
	}

	run, err := runner.GetRunner(runParams)
	if err != nil {
		return errors.WithStack(err)
	}

	err = run.Prepare()
	if err != nil {
		return errors.WithStack(err)
	}

	err = func() error {
		startTime := time.Now()

		messages.LaunchRunning.Notify(params.RequestContext, &butlerd.LaunchRunningNotification{})
		exitCode, err := interpretRunError(run.Run())
		messages.LaunchExited.Notify(params.RequestContext, &butlerd.LaunchExitedNotification{})
		if err != nil {
			return errors.WithStack(err)
		}

		runDuration := time.Since(startTime)
		err = params.RecordPlayTime(runDuration)
		if err != nil {
			return errors.WithStack(err)
		}

		if exitCode != 0 {
			var signedExitCode = int64(exitCode)
			if runtime.GOOS == "windows" {
				// Windows uses 32-bit unsigned integers as exit codes, although the
				// command interpreter treats them as signed. If a process fails
				// initialization, a Windows system error code may be returned.
				signedExitCode = int64(int32(signedExitCode))

				// The line above turns `4294967295` into -1
			}

			exeName := filepath.Base(params.FullTargetPath)
			msg := fmt.Sprintf("Exit code 0x%x (%d) for (%s)", uint32(exitCode), signedExitCode, exeName)
			consumer.Warnf(msg)

			if runDuration.Seconds() > 10 {
				consumer.Warnf("That's after running for %s, ignoring non-zero exit code", runDuration)
			} else {
				return errors.New(msg)
			}
		}

		return nil
	}()

	if err != nil {
		consumer.Errorf("Had error: %s", err.Error())
		if len(stderr.Lines()) == 0 {
			consumer.Errorf("No messages for standard error")
			consumer.Errorf("→ Standard error: empty")
		} else {
			consumer.Errorf("→ Standard error ================")
			for _, l := range stderr.Lines() {
				consumer.Errorf("  %s", l)
			}
			consumer.Errorf("=================================")
		}

		if len(stdout.Lines()) == 0 {
			consumer.Errorf("→ Standard output: empty")
		} else {
			consumer.Errorf("→ Standard output ===============")
			for _, l := range stdout.Lines() {
				consumer.Errorf("  %s", l)
			}
			consumer.Errorf("=================================")
		}
		consumer.Errorf("Relaying launch failure.")
		return errors.WithStack(err)
	}

	return nil
}

func configureTargetIfNeeded(params *launch.LauncherParams) error {
	if params.Candidate != nil {
		// already configured
		return nil
	}

	v, err := dash.Configure(params.FullTargetPath, &dash.ConfigureParams{
		Consumer: params.RequestContext.Consumer,
		Filter:   filtering.FilterPaths,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if len(v.Candidates) == 0 {
		return errors.Errorf("0 candidates after configure")
	}

	params.Candidate = v.Candidates[0]
	return nil
}

func fillPeInfoIfNeeded(params *launch.LauncherParams) error {
	c := params.Candidate
	if c == nil {
		// no candidate for some reason?
		return nil
	}

	if c.Flavor != dash.FlavorNativeWindows {
		// not an .exe, ignore
		return nil
	}

	var err error
	f, err := os.Open(params.FullTargetPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	params.PeInfo, err = pelican.Probe(f, &pelican.ProbeParams{
		Consumer: params.RequestContext.Consumer,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func interpretRunError(err error) (int, error) {
	if err != nil {
		if exitError, ok := AsExitError(err); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}

		return 127, err
	}

	return 0, nil
}

type causer interface {
	Cause() error
}

func AsExitError(err error) (*exec.ExitError, bool) {
	if err == nil {
		return nil, false
	}

	if se, ok := err.(causer); ok {
		return AsExitError(se.Cause())
	}

	if ee, ok := err.(*exec.ExitError); ok {
		return ee, true
	}

	return nil, false
}
