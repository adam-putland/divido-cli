package models

import "fmt"

type Platform struct {
	Release  *Release
	Services Services
}

func (plat *Platform) String() string {
	return fmt.Sprintf(" Plat name: %s\n latest version: %s\n URL: %s\n", plat.Release.Name, plat.Release.Version, plat.Release.URL)
}
