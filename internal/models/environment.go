package models

import (
	"fmt"
	"strings"
)

type Service struct {
	Name    string
	Version string
}

type Environment struct {
	EnvironmentConfig
	HelmChartVersion string
	Overrides        []byte
	Services         []byte
}

func (env Environment) Info() string {
	return fmt.Sprintf("helm version: %s\n overrides: %s", env.HelmChartVersion, env.Overrides)
}

func (env Environment) GetHCVersion() string {
	return fmt.Sprintf("v%s", strings.TrimSpace(env.HelmChartVersion))
}
