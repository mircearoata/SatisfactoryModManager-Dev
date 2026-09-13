package app

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
)

func (a *app) ProcessArguments(args []string) {
	a.argsMutex.Lock()
	defer a.argsMutex.Unlock()

	if !a.isFrontendReady {
		a.pendingArgs = append(a.pendingArgs, args)
		return
	}

	a.processArguments(args)
}

func (a *app) processPendingArgs() {
	a.argsMutex.Lock()
	defer a.argsMutex.Unlock()

	for _, args := range a.pendingArgs {
		a.processArguments(args)
	}

	a.pendingArgs = nil
}

func (a *app) processArguments(args []string) {
	if len(args) < 1 {
		return
	}
	if strings.HasPrefix(args[0], "smmanager://") {
		uri := args[0]
		err := a.handleURI(uri)
		if err != nil {
			slog.Error("failed to handle smmanager:// URI", slog.Any("error", err), slog.String("uri", uri))
		}
	} else {
		err := a.handleFile(args[0])
		if err != nil {
			slog.Error("failed to handle file", slog.Any("error", err), slog.String("path", args[0]))
		}
	}
	a.Show()
}

func (a *app) handleURI(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}
	switch u.Host {
	case "install":
		modID := u.Query().Get("modID")
		version := u.Query().Get("version")
		a.ExternalInstallMod(modID, version)
		return nil
	default:
		return fmt.Errorf("unknown URI action %s", u.Host)
	}
}

func (a *app) handleFile(path string) error {
	if strings.HasSuffix(path, ".smmprofile") {
		println(path)
		a.ExternalImportProfile(path)
		return nil
	}
	return fmt.Errorf("unknown file type %s", path)
}
