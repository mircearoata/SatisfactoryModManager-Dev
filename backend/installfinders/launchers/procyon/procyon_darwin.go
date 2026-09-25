package procyon

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/steam"
)

var procyonPreferencesRelativePaths = []string{
	filepath.Join("Library", "Group Containers", "group.com.italomandara.procyon", "Library", "Preferences", "group.com.italomandara.procyon.plist"),
	filepath.Join("Library", "Preferences", "group.com.italomandara.procyon.plist"),
}

type procyonPlist struct {
	SelectedBottle string `plist:"selectedBottle"`
}

func init() {
	launchers.Add("Procyon", procyon)
}

func procyon() ([]*common.Installation, []error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get user home directory: %w", err)}
	}

	var plistPath string
	for _, relativePath := range procyonPreferencesRelativePaths {
		path := filepath.Join(homeDir, relativePath)
		if _, err := os.Stat(path); err == nil {
			plistPath = path
			break
		}
	}
	if plistPath == "" {
		return nil, []error{fmt.Errorf("procyon not installed")}
	}

	plistBytes, err := os.ReadFile(plistPath)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to read Procyon preferences: %w", err)}
	}

	var prefs procyonPlist
	if _, err := plist.Unmarshal(plistBytes, &prefs); err != nil {
		return nil, []error{fmt.Errorf("failed to parse Procyon preferences: %w", err)}
	}

	if prefs.SelectedBottle == "" {
		return nil, nil
	}

	bottlePath, err := parseBottlePath(prefs.SelectedBottle, homeDir)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to parse Procyon bottle path: %w", err)}
	}

	realBottlePath, err := filepath.EvalSymlinks(bottlePath)
	if err != nil {
		slog.Warn("failed to evaluate symlink for Procyon bottle path, using original path", slog.String("path", bottlePath), slog.Any("error", err))
		realBottlePath = bottlePath
	}

	stat, err := os.Stat(realBottlePath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("Procyon bottle directory does not exist", slog.String("path", realBottlePath))
			return nil, nil
		}
		return nil, []error{fmt.Errorf("failed to stat Procyon bottle %s: %w", realBottlePath, err)}
	}

	if !stat.IsDir() {
		slog.Warn("Procyon bottle path is not a directory", slog.String("path", realBottlePath))
		return nil, nil
	}

	return steam.FindInstallationsWine(realBottlePath, "Procyon", nil)
}

func parseBottlePath(raw string, homeDir string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", nil
	}

	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
		unescaped, err := url.PathUnescape(path)
		if err != nil {
			return "", fmt.Errorf("failed to unescape bottle URL %s: %w", raw, err)
		}
		path = unescaped
	}

	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, strings.TrimPrefix(path, "~"))
	}

	return filepath.Clean(path), nil
}
