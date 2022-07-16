package models

import "fmt"

type Platform struct {
	Release
	Services []byte
}

func (plat *Platform) String() string {
	return fmt.Sprintf(" Plat name: %s\n latest version: %s\n URL: %s\n Services: \n%s", plat.Name, plat.Version, plat.URL, plat.Services)
}
