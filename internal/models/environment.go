package models

import (
	"fmt"
	"strings"
)

type Environment struct {
	EnvironmentConfig
	HelmChartVersion string
	Overrides        Services
	Services         Services
}

func (env Environment) String() string {
	return fmt.Sprintf("Name: %s\nhelm version: %s\n", env.Name, env.HelmChartVersion)
}

func (env Environment) GetHCVersion() string {
	return fmt.Sprintf("v%s", strings.TrimSpace(env.HelmChartVersion))
}
