package models

import (
	"fmt"
	"github.com/mgutz/ansi"
	"strings"
)

type ServiceUpdated struct {
	Service    *Service
	NewVersion string
}

type Comparer struct {
	InitialVersion string
	FinalVersion   string
	Insert         Services
	Deleted        Services
	Changed        map[string]*ServiceUpdated
	DisableColor   bool
}

func Compare(plat1, plat2 *Platform) *Comparer {

	changed := make(map[string]*ServiceUpdated)

	var comparer Comparer

	var services1 Services
	var services2 Services
	if plat1.Release.Date.After(plat2.Release.Date) {
		services2 = make(Services, len(plat1.Services))
		for k, v := range plat1.Services {
			services2[k] = v
		}
		services1 = make(Services, len(plat2.Services))
		for k, v := range plat2.Services {
			services1[k] = v
		}

		comparer.InitialVersion = plat2.Release.Version
		comparer.FinalVersion = plat1.Release.Version
	} else {
		services1 = make(Services, len(plat1.Services))
		for k, v := range plat1.Services {
			services1[k] = v
		}
		services2 = make(Services, len(plat2.Services))
		for k, v := range plat2.Services {
			services2[k] = v
		}
		comparer.InitialVersion = plat1.Release.Version
		comparer.FinalVersion = plat2.Release.Version
	}

	for s, service1 := range services1 {
		if service2, ok := services2[s]; ok {
			if service1.Version != service2.Version {
				changed[service1.HLMName] = &ServiceUpdated{
					Service:    service1,
					NewVersion: service2.Version,
				}
			}
			delete(services1, s)
			delete(services2, s)
			continue
		}

	}

	comparer.Insert = services2
	comparer.Deleted = services1
	comparer.Changed = changed
	return &comparer

}

func (c *Comparer) String() string {

	var builder strings.Builder

	fmt.Fprintf(&builder, "Base Helm Chart Change: %s -> %s\n\n", c.InitialVersion, c.FinalVersion)

	if len(c.Changed) > 0 {
		builder.WriteString("Service Versions Changes:\n")
		for k, ch := range c.Changed {
			fmt.Fprintf(&builder, "- %s: %s -> %s\n", k, c.MakeDiffText(ch.Service.Version, "red"), c.MakeDiffText(ch.NewVersion, "green"))
		}
	}

	if len(c.Insert) > 0 {
		builder.WriteString("Service Versions Included:\n")
		for k, i := range c.Insert {
			fmt.Fprintf(&builder, " %s: %s\n", k, c.MakeDiffText(i.Version, "green"))
		}
	}

	if len(c.Deleted) > 0 {
		builder.WriteString("Service Versions Excluded:\n")
		for k, i := range c.Insert {
			fmt.Fprintf(&builder, " %s: %s\n", k, c.MakeDiffText(i.Version, "red"))
		}
	}

	return builder.String()
}

func (c *Comparer) MakeDiffText(text, color string) string {
	if c.DisableColor {
		return text
	}
	return ansi.Color(text, color)
}
