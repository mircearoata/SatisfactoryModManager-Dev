package wineskin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/epic"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/steam"
)

func init() {
	launchers.Add("Wineskin", wineskin)
}

func wineskin() ([]*common.Installation, []error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get user home directory: %w", err)}
	}

	dedicatedDirs := []string{
		filepath.Join(homeDir, "Applications", "Wineskin"),
		"/Applications/Wineskin",
	}

	generalDirs := []string{
		filepath.Join(homeDir, "Applications"),
		"/Applications",
	}

	var rawWrappers []string
	anyDedicatedDirExists := false

	for _, dir := range dedicatedDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		anyDedicatedDirExists = true
		for _, entry := range entries {
			if !strings.HasSuffix(strings.ToLower(entry.Name()), ".app") {
				continue
			}
			if strings.EqualFold(entry.Name(), "Wineskin Winery.app") {
				continue
			}
			rawWrappers = append(rawWrappers, filepath.Join(dir, entry.Name()))
		}
	}

	for _, dir := range generalDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !strings.HasSuffix(strings.ToLower(entry.Name()), ".app") {
				continue
			}
			if strings.EqualFold(entry.Name(), "Wineskin Winery.app") {
				continue
			}
			appPath := filepath.Join(dir, entry.Name())
			markerPath := filepath.Join(appPath, "Contents", "Wineskin.app")
			if stat, err := os.Stat(markerPath); err == nil && stat.IsDir() {
				rawWrappers = append(rawWrappers, appPath)
			}
		}
	}

	wrappers := common.UniquePaths(rawWrappers)

	if !anyDedicatedDirExists && len(wrappers) == 0 {
		return nil, []error{fmt.Errorf("wineskin not installed")}
	}

	installations := make([]*common.Installation, 0)
	errors := make([]error, 0)

	for _, wrapper := range wrappers {
		stat, err := os.Stat(wrapper)
		if err != nil {
			if !os.IsNotExist(err) {
				errors = append(errors, fmt.Errorf("failed to stat Wineskin wrapper %s: %w", wrapper, err))
			}
			continue
		}
		if !stat.IsDir() {
			continue
		}

		prefixPath := filepath.Join(wrapper, "Contents", "SharedSupport", "prefix")

		search := func(finder func(winePrefix string, launcher string, launchPath []string) ([]*common.Installation, []error), storeName string) {
			bottleInstalls, bottleErrs := finder(prefixPath, fmt.Sprintf("Wineskin - %s", storeName), nil)
			installations = append(installations, bottleInstalls...)
			if bottleErrs != nil {
				errors = append(errors, bottleErrs...)
			}
		}

		search(steam.FindInstallationsWine, "Steam")
		search(epic.FindInstallationsWine, "Epic")
	}

	return installations, errors
}
