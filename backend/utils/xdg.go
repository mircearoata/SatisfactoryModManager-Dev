package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func UserDataDir() (string, error) {
	dataDir, ok := os.LookupEnv("XDG_DATA_HOME")
	if !ok {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home dir: %w", err)
		}
		dataDir = filepath.Join(homeDir, ".local", "share")
	}
	return dataDir, nil
}
