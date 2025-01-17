package prereqs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/itchio/butler/cmd/dl"
	"github.com/itchio/butler/comm"
	"github.com/itchio/butler/mansion"
	"github.com/itchio/butler/redist"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

func Test(ctx *mansion.Context, prereqs []string) error {
	comm.Opf("Fetching registry...")

	baseURL := "https://dl.itch.ovh/itch-redists"

	infoURL := fmt.Sprintf("%s/info.json?t=%d", baseURL, time.Now().Unix())
	res, err := http.Get(infoURL)
	if err != nil {
		return errors.WithStack(err)
	}

	if res.StatusCode != 200 {
		return errors.Errorf("While getting redist registry, got HTTP %d", res.StatusCode)
	}

	dec := json.NewDecoder(res.Body)

	registry := &redist.RedistRegistry{}
	err = dec.Decode(registry)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(prereqs) == 0 {
		comm.Logf("")
		comm.Statf("No prereqs specified, here are those we know about: \n")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetColWidth(60)
		table.SetHeader([]string{"Name", "Arch", "Description", "Version"})

		var entries []*NamedRedistEntry
		for name, v := range registry.Entries {
			entries = append(entries, &NamedRedistEntry{name, v})
		}
		sort.Stable(&entriesByArch{entries})
		sort.Stable(&entriesByName{entries})
		for _, e := range entries {
			info := e.entry
			table.Append([]string{e.name, info.Arch, info.FullName, info.Version})
		}
		table.Render()
		return nil
	}

	if len(prereqs) == 1 && prereqs[0] == "all" {
		prereqs = nil
		for name := range registry.Entries {
			prereqs = append(prereqs, name)
		}
	}

	comm.Logf("Testing out prereqs %s", strings.Join(prereqs, ", "))

	plan := &PrereqPlan{}

	tempDir := filepath.Join(os.TempDir(), "butler-test-prereqs")
	comm.Logf("Working in %s", tempDir)
	comm.Logf("(This helps not having to re-download the prereqs between runs, but feel free to wipe it)")

	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, name := range prereqs {
		info, ok := registry.Entries[name]
		if !ok {
			comm.Warnf("Unknown prereq %s, skipping", name)
			continue
		}

		block := info.Windows
		if block == nil {
			return errors.Errorf("No windows block for prereq %s", name)
		}

		comm.Opf("Downloading prereq %s", name)

		workDir := filepath.Join(tempDir, name)
		err = os.MkdirAll(workDir, 0755)
		if err != nil {
			return errors.WithStack(err)
		}

		task := &PrereqTask{
			Info:    info,
			Name:    name,
			WorkDir: workDir,
		}

		url := fmt.Sprintf("%s/%s/%s", baseURL, name, block.Command)
		dest := filepath.Join(workDir, block.Command)
		_, err = dl.Do(ctx, url, dest)
		if err != nil {
			comm.Logf("Could not download prereq %s", name)
			return errors.WithStack(err)
		}

		plan.Tasks = append(plan.Tasks, task)
	}

	planPath := filepath.Join(tempDir, "butler_install_plan.json")
	comm.Logf("Writing plan to %s", planPath)

	planContents, err := json.Marshal(plan)
	if err != nil {
		return errors.WithStack(err)
	}

	err = ioutil.WriteFile(planPath, planContents, 0644)
	if err != nil {
		return errors.WithStack(err)
	}

	comm.Opf("Handing off to install-prereqs...")

	err = Install(ctx, planPath, "")
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
