package models

import "fmt"

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
