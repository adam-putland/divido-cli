package internal

type Config struct {
	Platforms []Platform
}

type Platform struct {
	Name      string
	HelmChartRepo string `mapstructure:"hlm"`
	Envs      []Environment
}

type Environment struct {
	Name      string
	Repo      string
	ChartPath string `mapstructure:",omitempty"`
}

func (c Config) ListPlatform() []string {
	platforms := make([]string, 0, len(c.Platforms))
	for _, platform := range c.Platforms {
		platforms = append(platforms, platform.Name)
	}
	return platforms
}

func (c Config) ListEnvironments(platformIndex int) []string {
	p := c.GetPlatform(platformIndex)

	if p == nil {
		return nil
	}

	environments := make([]string, 0, len(p.Envs))

	for _, env := range p.Envs {
		environments = append(environments, env.Name)
	}

	return environments
}

func (c Config) GetPlatform(platformIndex int) *Platform {
	if platformIndex < 0 {
		return nil
	}
	return &c.Platforms[platformIndex]
}

func (p *Platform) GetEnvironment(envIndex int) *Environment {
	return &p.Envs[envIndex]
}
