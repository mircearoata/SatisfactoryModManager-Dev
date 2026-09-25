package cxpatcher

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/epic"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/steam"
)

var (
	cxpatcherDefaultBottlesRelativePath = "CXPBottles"
	crossOverConfSubPath                = filepath.Join("Contents", "SharedSupport", "CrossOver", "etc", "CrossOver.conf")
)

func init() {
	launchers.Add("CXPatcher", cxpatcher)
}

func cxpatcher() ([]*common.Installation, []error) {
	bottlesPaths, err := getCXPatcherBottlesPaths()
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get CXPatcher bottles paths: %w", err)}
	}

	if len(bottlesPaths) == 0 {
		return nil, []error{fmt.Errorf("cxpatcher not installed")}
	}

	installations := make([]*common.Installation, 0)
	errors := make([]error, 0)

	for _, bottlesPath := range bottlesPaths {
		bottles, err := os.ReadDir(bottlesPath)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to list CXPatcher bottles in %s: %w", bottlesPath, err))
			continue
		}

		for _, bottle := range bottles {
			bottleRoot := filepath.Join(bottlesPath, bottle.Name())
			realBottleRoot, err := filepath.EvalSymlinks(bottleRoot)
			if err != nil {
				slog.Warn("failed to evaluate symlink for bottle path, using original path", slog.String("path", bottleRoot), slog.Any("error", err))
				realBottleRoot = bottleRoot
			}

			stat, err := os.Stat(realBottleRoot)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to stat CXPatcher bottle %s: %w", realBottleRoot, err))
				continue
			}
			if !stat.IsDir() {
				continue
			}

			search := func(finder func(winePrefix string, launcher string, launchPath []string) ([]*common.Installation, []error)) {
				bottleInstalls, bottleErrs := finder(realBottleRoot, "CXPatcher", nil)
				installations = append(installations, bottleInstalls...)
				if bottleErrs != nil {
					errors = append(errors, bottleErrs...)
				}
			}
			search(steam.FindInstallationsWine)
			search(epic.FindInstallationsWine)
		}
	}

	return installations, errors
}

func getCXPatcherBottlesPaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	currentUser, err := user.Current()
	username := ""
	if err != nil {
		slog.Warn("failed to get current user", slog.Any("error", err))
	} else {
		username = currentUser.Username
	}

	rawPaths := []string{
		filepath.Join(homeDir, cxpatcherDefaultBottlesRelativePath),
	}

	appDirs := []string{
		"/Applications",
		filepath.Join(homeDir, "Applications"),
	}
	appNames := []string{
		"CrossOver.app",
		"CrossOver_patched.app",
	}

	for _, appDir := range appDirs {
		for _, appName := range appNames {
			confPath := filepath.Join(appDir, appName, crossOverConfSubPath)
			if confBytes, err := os.ReadFile(confPath); err == nil {
				confBottlePaths := parseCrossOverConf(confBytes)
				rawPaths = append(rawPaths, confBottlePaths...)
			}
		}
	}

	var existingDirs []string
	for _, raw := range rawPaths {
		expanded := expandBottlePath(raw, homeDir, username)
		if expanded == "" {
			continue
		}
		realPath, err := filepath.EvalSymlinks(expanded)
		if err != nil {
			slog.Warn("failed to evaluate symlink for bottle path, using original path", slog.String("path", expanded), slog.Any("error", err))
			realPath = expanded
		}
		if stat, err := os.Stat(realPath); err == nil && stat.IsDir() {
			existingDirs = append(existingDirs, realPath)
		}
	}

	return deduplicatePaths(existingDirs), nil
}

func parseCrossOverConf(content []byte) []string {
	cfg, err := ini.Load(content)
	if err != nil {
		slog.Error("failed to parse CrossOver.conf", slog.Any("error", err))
		return nil
	}

	sec := cfg.Section("EnvironmentVariables")
	if sec == nil {
		return nil
	}

	rawVal := ""
	if sec.HasKey("CX_BOTTLE_PATH") {
		rawVal = sec.Key("CX_BOTTLE_PATH").String()
	} else if sec.HasKey(`"CX_BOTTLE_PATH"`) {
		rawVal = sec.Key(`"CX_BOTTLE_PATH"`).String()
	}

	parts := strings.Split(rawVal, ":")
	var paths []string
	for _, p := range parts {
		trimmed := strings.Trim(p, `"' `)
		if trimmed != "" {
			paths = append(paths, trimmed)
		}
	}
	return paths
}

func expandBottlePath(rawPath, homeDir, username string) string {
	res := strings.ReplaceAll(rawPath, "${USER}", username)
	res = strings.ReplaceAll(res, "$USER", username)
	res = strings.ReplaceAll(res, "${HOME}", homeDir)
	res = strings.ReplaceAll(res, "$HOME", homeDir)

	if strings.HasPrefix(res, "~/") {
		res = filepath.Join(homeDir, res[2:])
	} else if res == "~" {
		res = homeDir
	}

	return filepath.Clean(res)
}

func deduplicatePaths(paths []string) []string {
	var unique []string
	for _, p := range paths {
		cleaned := filepath.Clean(p)
		exists := false
		for _, u := range unique {
			if common.OsPathEqual(cleaned, u) {
				exists = true
				break
			}
		}
		if !exists {
			unique = append(unique, cleaned)
		}
	}
	return unique
}
