package buse

import (
	"time"

	"github.com/itchio/butler/cmd/launch/manifest"
	"github.com/itchio/butler/configurator"
	"github.com/itchio/butler/installer/bfs"
	itchio "github.com/itchio/go-itchio"
)

// must be kept in sync with clients, see for example
// https://github.com/itchio/node-butler

//----------------------------------------------------------------------
// Version
//----------------------------------------------------------------------

// Retrieves the version of the butler instance the client
// is connected to.
//
// This endpoint is meant to gather information when reporting
// issues, rather than feature sniffing. Conforming clients should
// automatically download new versions of butler, see [Updating](#updating).
//
// @name Version.Get
// @category Utilities
// @tags Offline
type VersionGetParams struct{}

type VersionGetResult struct {
	// Something short, like `v8.0.0`
	Version string `json:"version"`

	// Something long, like `v8.0.0, built on Aug 27 2017 @ 01:13:55, ref d833cc0aeea81c236c81dffb27bc18b2b8d8b290`
	VersionString string `json:"versionString"`
}

//----------------------------------------------------------------------
// Game
//----------------------------------------------------------------------

// Finds uploads compatible with the current runtime, for a given game
//
// @name Game.FindUploads
// @category Install
type GameFindUploadsParams struct {
	// Which game to find uploads for
	Game *itchio.Game `json:"game"`
	// The credentials to use to list uploads
	Credentials *GameCredentials `json:"credentials"`
}

type GameFindUploadsResult struct {
	// A list of uploads that were found to be compatible.
	Uploads []*itchio.Upload `json:"uploads"`
}

//----------------------------------------------------------------------
// Operation
//----------------------------------------------------------------------

type Operation string

var (
	OperationInstall   Operation = "install"
	OperationUninstall Operation = "uninstall"
)

// Start a new operation (installing or uninstalling).
//
// @name Operation.Start
// @category Install
// @tags Cancellable
type OperationStartParams struct {
	// A UUID, generated by the client, used for referring to the
	// task when cancelling it, for instance.
	ID string `json:"id"`

	// A folder that butler can use to store temporary files, like
	// partial downloads, checkpoint files, etc.
	StagingFolder string `json:"stagingFolder"`

	// Which operation to perform
	Operation Operation `json:"operation"`

	// Must be set if Operation is `install`
	InstallParams *InstallParams `json:"installParams,omitempty"`

	// Must be set if Operation is `uninstall`
	UninstallParams *UninstallParams `json:"uninstallParams,omitempty"`
}

type OperationStartResult struct{}

// Attempt to gracefully cancel an ongoing operation.
//
// @name Operation.Cancel
// @category Install
type OperationCancelParams struct {
	// The UUID of the task to cancel, as passed to [Operation.Start](#operationstart-request)
	ID string `json:"id"`
}

type OperationCancelResult struct{}

// InstallParams contains all the parameters needed to perform
// an installation for a game.
// @kind type
// @category Install
type InstallParams struct {
	// Which game to install
	Game *itchio.Game `json:"game"`

	// An absolute path where to install the game
	InstallFolder string `json:"installFolder"`

	// Which upload to install
	// @optional
	Upload *itchio.Upload `json:"upload"`

	// Which build to install
	// @optional
	Build *itchio.Build `json:"build"`

	// Which credentials to use to install the game
	Credentials *GameCredentials `json:"credentials"`

	// If true, do not run windows installers, just extract
	// whatever to the install folder.
	// @optional
	IgnoreInstallers bool `json:"ignoreInstallers,omitempty"`
}

// UninstallParams contains all the parameters needed to perform
// an uninstallation for a game.
// @kind type
// @category Install
type UninstallParams struct {
	// Absolute path of the folder butler should uninstall
	InstallFolder string `json:"installFolder"`
}

// GameCredentials contains all the credentials required to make API requests
// including the download key if any
// @category General
type GameCredentials struct {
	// Defaults to `https://itch.io`
	Server string `json:"server"`
	// A valid itch.io API key
	APIKey string `json:"apiKey"`
	// A download key identifier, or 0 if no download key is available
	DownloadKey int64 `json:"downloadKey"`
}

// Asks the user to pick between multiple available uploads
//
// @category Install
// @tags Dialog
type PickUploadParams struct {
	// An array of upload objects to choose from
	Uploads []*itchio.Upload `json:"uploads"`
}

type PickUploadResult struct {
	// The index (in the original array) of the upload that was picked,
	// or a negative value to cancel.
	Index int64 `json:"index"`
}

// Retrieves existing receipt information for an install
//
// @category Install
// @tags Deprecated
type GetReceiptParams struct {
	// muffin
}

type GetReceiptResult struct {
	Receipt *bfs.Receipt `json:"receipt"`
}

// Sent periodically to inform on the current state an operation.
//
// @name Operation.Progress
// @category Install
type OperationProgressNotification struct {
	// An overall progress value between 0 and 1
	Progress float64 `json:"progress"`
	// Estimated completion time for the operation, in seconds (floating)
	ETA float64 `json:"eta"`
	// Network bandwidth used, in bytes per second (floating)
	BPS float64 `json:"bps"`
}

type TaskReason string

const (
	TaskReasonInstall   TaskReason = "install"
	TaskReasonUninstall TaskReason = "uninstall"
)

type TaskType string

const (
	TaskTypeDownload  TaskType = "download"
	TaskTypeInstall   TaskType = "install"
	TaskTypeUninstall TaskType = "uninstall"
	TaskTypeUpdate    TaskType = "update"
	TaskTypeHeal      TaskType = "heal"
)

// Each operation is made up of one or more tasks. This notification
// is sent whenever a task starts for an operation.
//
// @category Install
type TaskStartedNotification struct {
	// Why this task was started
	Reason TaskReason `json:"reason"`
	// Is this task a download? An install?
	Type TaskType `json:"type"`
	// The game this task is dealing with
	Game *itchio.Game `json:"game"`
	// The upload this task is dealing with
	Upload *itchio.Upload `json:"upload"`
	// The build this task is dealing with (if any)
	Build *itchio.Build `json:"build,omitempty"`
	// Total size in bytes
	TotalSize int64 `json:"totalSize,omitempty"`
}

// Sent whenever a task succeeds for an operation.
//
// @category Install
type TaskSucceededNotification struct {
	Type TaskType `json:"type"`
	// If the task installed something, then this contains
	// info about the game, upload, build that were installed
	InstallResult *InstallResult `json:"installResult,omitempty"`
}

type OperationResult struct{}

// @category Install
// @kind type
type InstallResult struct {
	Game   *itchio.Game   `json:"game"`
	Upload *itchio.Upload `json:"upload"`
	Build  *itchio.Build  `json:"build"`
	// TODO: verdict ?
}

//----------------------------------------------------------------------
// CheckUpdate
//----------------------------------------------------------------------

// Looks for one or more game updates.
//
// @category Update
type CheckUpdateParams struct {
	// A list of items, each of it will be checked for updates
	Items []*CheckUpdateItem `json:"items"`
}

// @category Update
type CheckUpdateItem struct {
	// An UUID generated by the client, which allows it to map back the
	// results to its own items.
	ItemID string `json:"itemId"`
	// Timestamp of the last successful install operation
	InstalledAt string `json:"installedAt"`
	// Game for which to look for an update
	Game *itchio.Game `json:"game"`
	// Currently installed upload
	Upload *itchio.Upload `json:"upload"`
	// Currently installed build
	Build *itchio.Build `json:"build,omitempty"`
	// Credentials to use to list uploads
	Credentials *GameCredentials `json:"credentials"`
}

type CheckUpdateResult struct {
	// Any updates found (might be empty)
	Updates []*GameUpdate `json:"updates"`
	// Warnings messages logged while looking for updates
	Warnings []string `json:"warnings"`
}

// Sent while CheckUpdate is still running, every time butler
// finds an update for a game. Can be safely ignored if displaying
// updates as they are found is not a requirement for the client.
//
// @category Update
// @tags Optional
type GameUpdateAvailableNotification struct {
	Update *GameUpdate `json:"update"`
}

// Describes an available update for a particular game install.
//
// @category Update
type GameUpdate struct {
	// Identifier originally passed in CheckUpdateItem
	ItemID string `json:"itemId"`
	// Game we found an update for
	Game *itchio.Game `json:"game"`
	// Upload to be installed
	Upload *itchio.Upload `json:"upload"`
	// Build to be installed (may be nil)
	Build *itchio.Build `json:"build"`
}

//----------------------------------------------------------------------
// Launch
//----------------------------------------------------------------------

// @category Launch
type LaunchParams struct {
	InstallFolder string                `json:"installFolder"`
	Game          *itchio.Game          `json:"game"`
	Upload        *itchio.Upload        `json:"upload"`
	Build         *itchio.Build         `json:"build"`
	Verdict       *configurator.Verdict `json:"verdict"`

	PrereqsDir   string `json:"prereqsDir"`
	ForcePrereqs bool   `json:"forcePrereqs,omitempty"`

	Sandbox bool `json:"sandbox,omitempty"`

	// Used for subkeying
	Credentials *GameCredentials `json:"credentials"`
}

type LaunchResult struct {
}

// Sent when the game is configured, prerequisites are installed
// sandbox is set up (if enabled), and the game is actually running.
//
// @category Launch
type LaunchRunningNotification struct{}

// Sent when the game has actually exited.
//
// @category Launch
type LaunchExitedNotification struct{}

// Pick a manifest action to launch, see [itch app manifests](https://itch.io/docs/itch/integrating/manifest.html)
//
// @tags Dialogs
// @category Launch
type PickManifestActionParams struct {
	Actions []*manifest.Action `json:"actions"`
}

type PickManifestActionResult struct {
	Name string `json:"name"`
}

// Ask the client to perform a shell launch, ie. open an item
// with the operating system's default handler (File explorer)
//
// @category Launch
type ShellLaunchParams struct {
	ItemPath string `json:"itemPath"`
}

type ShellLaunchResult struct {
}

// Ask the client to perform an HTML launch, ie. open an HTML5
// game, ideally in an embedded browser.
//
// @category Launch
type HTMLLaunchParams struct {
	RootFolder string `json:"rootFolder"`
	IndexPath  string `json:"indexPath"`

	Args []string          `json:"args"`
	Env  map[string]string `json:"env"`
}

type HTMLLaunchResult struct {
}

// Ask the client to perform an URL launch, ie. open an address
// with the system browser or appropriate.
//
// @category Launch
type URLLaunchParams struct {
	URL string `json:"url"`
}

type URLLaunchResult struct{}

// Ask the client to save verdict information after a reconfiguration.
//
// @category Launch
// @tags Deprecated
type SaveVerdictParams struct {
	Verdict *configurator.Verdict `json:"verdict"`
}
type SaveVerdictResult struct{}

// Ask the user to allow sandbox setup. Will be followed by
// a UAC prompt (on Windows) or a pkexec dialog (on Linux) if
// the user allows.
//
// @category Launch
// @tags Dialogs
type AllowSandboxSetupParams struct{}

type AllowSandboxSetupResult struct {
	Allow bool `json:"allow"`
}

// Sent when some prerequisites are about to be installed.
//
// @category Launch
type PrereqsStartedNotification struct {
	Tasks map[string]*PrereqTask `json:"tasks"`
}

// @category Launch
type PrereqTask struct {
	FullName string `json:"fullName"`
	Order    int    `json:"order"`
}

type PrereqStatus string

const (
	PrereqStatusPending     PrereqStatus = "pending"
	PrereqStatusDownloading PrereqStatus = "downloading"
	PrereqStatusReady       PrereqStatus = "ready"
	PrereqStatusInstalling  PrereqStatus = "installing"
	PrereqStatusDone        PrereqStatus = "done"
)

// @category Launch
type PrereqsTaskStateNotification struct {
	Name     string       `json:"name"`
	Status   PrereqStatus `json:"status"`
	Progress float64      `json:"progress"`
	ETA      float64      `json:"eta"`
	BPS      float64      `json:"bps"`
}

// @category Launch
type PrereqsEndedNotification struct {
}

// @category Launch
type PrereqsFailedParams struct {
	Error      string `json:"error"`
	ErrorStack string `json:"errorStack"`
}

type PrereqsFailedResult struct {
	Continue bool `json:"continue"`
}

//----------------------------------------------------------------------
// CleanDownloads
//----------------------------------------------------------------------

// Look for folders we can clean up in various download folders.
// This finds anything that doesn't correspond to any current downloads
// we know about.
//
// @name CleanDownloads.Search
// @category Clean Downloads
type CleanDownloadsSearchParams struct {
	// A list of folders to scan for potential subfolders to clean up
	Roots []string `json:"roots"`
	// A list of subfolders to not consider when cleaning
	// (staging folders for in-progress downloads)
	Whitelist []string `json:"whitelist"`
}

// @category Clean Downloads
type CleanDownloadsSearchResult struct {
	Entries []*CleanDownloadsEntry `json:"entries"`
}

// @category Clean Downloads
type CleanDownloadsEntry struct {
	// The complete path of the file or folder we intend to remove
	Path string `json:"path"`
	// The size of the folder or file, in bytes
	Size int64 `json:"size"`
}

// Remove the specified entries from disk, freeing up disk space.
//
// @name CleanDownloads.Apply
// @category Clean Downloads
type CleanDownloadsApplyParams struct {
	Entries []*CleanDownloadsEntry `json:"entries"`
}

// @category Clean Downloads
type CleanDownloadsApplyResult struct{}

//----------------------------------------------------------------------
// Misc.
//----------------------------------------------------------------------

// Sent any time butler needs to send a log message. The client should
// relay them in their own stdout / stderr, and collect them so they
// can be part of an issue report if something goes wrong.
//
// Log
type LogNotification struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

// @name Test.DoubleTwice
// @category Test
type TestDoubleTwiceParams struct {
	Number int64 `json:"number"`
}

// @category Test
type TestDoubleTwiceResult struct {
	Number int64 `json:"number"`
}

// @name Test.Double
// @category Test
type TestDoubleParams struct {
	Number int64 `json:"number"`
}

// Result for Test.Double
type TestDoubleResult struct {
	Number int64 `json:"number"`
}

const (
	CodeOperationCancelled = 499
	CodeOperationAborted   = 410
)

// Dates

func FromDateTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func ToDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
