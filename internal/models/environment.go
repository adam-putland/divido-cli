package models

import (
	"fmt"
	"strings"
)

type Environment struct {
	EnvironmentConfig
	HelmChartVersion string
	Overrides        []byte
	Services         []byte
}

func (env Environment) String() string {
	return fmt.Sprintf(" helm version: %s\n overrides: \n%s", env.HelmChartVersion, env.Overrides)
}

func (env Environment) GetHCVersion() string {
	return fmt.Sprintf("v%s", strings.TrimSpace(env.HelmChartVersion))
}
