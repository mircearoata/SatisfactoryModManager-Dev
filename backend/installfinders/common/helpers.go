package common

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func OsPathEqual(path1, path2 string) bool {
	path1 = realPath(path1)
	path2 = realPath(path2)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(path1, path2)
	}
	return path1 == path2
}

func UniquePaths(paths []string) []string {
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		exists := false
		for _, u := range unique {
			path1 := realPath(path)
			path2 := realPath(u)
			if OsPathEqual(path1, path2) {
				exists = true
				break
			}
		}
		if !exists {
			unique = append(unique, path)
		}
	}
	return unique
}

func realPath(path string) string {
	newPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("failed to evaluate symlink, using original path", slog.String("path", path), slog.Any("error", err))
		}
		return path
	}
	return newPath
}

func FindAll(finders ...InstallFinderFunc) ([]*Installation, []error) {
	installs := make([]*Installation, 0)
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
