package launchers

import (
	"cmp"
	"maps"
	"slices"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
)

type finderRegistration struct {
	finder common.InstallFinderFunc
	order  int
}

type InstallFinder struct {
	ID     string
	Finder common.InstallFinderFunc
}

var finders map[string]finderRegistration

func Add(id string, f common.InstallFinderFunc) {
	if finders == nil {
		finders = make(map[string]finderRegistration)
	}
	if _, ok := finders[id]; ok {
		panic("launcher already registered")
	}
	finders[id] = finderRegistration{finder: f, order: 0}
}

func AddOrdered(id string, f common.InstallFinderFunc, order int) {
	if finders == nil {
		finders = make(map[string]finderRegistration)
	}
	if _, ok := finders[id]; ok {
		panic("launcher already registered")
	}
	finders[id] = finderRegistration{finder: f, order: order}
}

func GetInstallFinders() []InstallFinder {
	orderedFinders := map[int][]InstallFinder{}
	for id, reg := range finders {
		orderedFinders[reg.order] = append(orderedFinders[reg.order], InstallFinder{ID: id, Finder: reg.finder})
	}
	orderKeys := slices.Sorted(maps.Keys(orderedFinders))

	registrations := make([]InstallFinder, 0, len(finders))
	for _, order := range orderKeys {
		orderRegistrations := orderedFinders[order]
		slices.SortFunc(orderRegistrations, func(a, b InstallFinder) int {
			return cmp.Compare(a.ID, b.ID)
		})
		registrations = append(registrations, orderRegistrations...)
	}
	return registrations
}
