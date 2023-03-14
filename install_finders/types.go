package install_finders

import (
	"path/filepath"
	"runtime"
	"strings"
)

type GameBranch string

var (
	BRANCH_EARLY_ACCESS GameBranch = "Early Access"
	BRANCH_EXPERIMENTAL GameBranch = "Experimental"
)

type Installation struct {
	Path       string     `json:"path"`
	Version    int        `json:"version"`
	Branch     GameBranch `json:"branch"`
	Launcher   string     `json:"launcher"`
	LaunchPath []string   `json:"launchPath"`
}

type InstallFindError struct {
	Inner error  `json:"cause"`
	Path  string `json:"path"`
}

func (e InstallFindError) Error() string {
	return e.Path + ": " + e.Inner.Error()
}

func (e InstallFindError) Causes() error {
	return e.Inner
}

type InstallFinderFunc func() ([]*Installation, []error)

func OsPathEqual(path1, path2 string) bool {
	path1 = filepath.Clean(path1)
	path2 = filepath.Clean(path2)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(path1, path2)
	}
	return path1 == path2
}

func FindAll(finders ...InstallFinderFunc) ([]*Installation, []error) {
	var installs []*Installation
	var errors []error
	for _, finder := range finders {
		foundInstalls, foundErrors := finder()
		for _, install := range foundInstalls {
			existing := false
			for i := range installs {
				if OsPathEqual(installs[i].Path, install.Path) {
					existing = true
					break
				}
			}
			if !existing {
				installs = append(installs, install)
			}
		}
		errors = append(errors, foundErrors...)
	}
	return installs, errors
}
