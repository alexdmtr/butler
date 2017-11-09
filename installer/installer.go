package installer

import (
	"github.com/itchio/butler/archive"
	"github.com/itchio/butler/installer/bfs"
	"github.com/itchio/wharf/state"
)

type Manager interface {
	Install(params *InstallParams) (*InstallResult, error)
	Uninstall(params *UninstallParams) error
	Name() string
}

type InstallParams struct {
	// An archive file, .exe setup file, .dmg file etc.
	SourcePath string

	// The existing receipt, if any
	ReceiptIn *bfs.Receipt

	// Where the item should be installed
	InstallFolderPath string

	Consumer *state.Consumer

	ArchiveListResult archive.ListResult
}

type UninstallParams struct {
	InstallPath string
}

type InstallResult struct {
	// Files is a list of paths, relative to the install folder
	Files []string
}
