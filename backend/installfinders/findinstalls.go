package installfinders

import (
	"log/slog"
	"strings"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers"

	_ "github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/all" // register all launchers
)

func FindInstallations() ([]*common.Installation, []error) {
	registrations := launchers.GetInstallFinders()

	names := make([]string, 0, len(registrations))
	finders := make([]common.InstallFinderFunc, 0, len(registrations))
	for _, registration := range registrations {
		names = append(names, registration.ID)
		finders = append(finders, registration.Finder)
	}

	slog.Debug("finding installations", slog.String("launchers", strings.Join(names, ",")))

	return common.FindAll(finders...)
}
