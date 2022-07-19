package models

type Config struct {
	Platforms []PlatformConfig
	Github    GithubConfig
	Services  map[string]string
}

type GithubConfig struct {
	Org                      string
	AuthorName               string
	AuthorEmail              string
	MainBranch               string
	Message                  string
	PreCommitMessage         string
	CommitMessageBumpHc      string
	CommitMessageBumpService string
}

type PlatformConfig struct {
	Name          string
	HelmChartRepo string `mapstructure:"hlm"`
	Envs          []EnvironmentConfig
	DirectCommit  bool
}

type ServicesConfig struct {
	Repo          string
	HelmChartRepo string `mapstructure:"hlmName"`
}

type EnvironmentConfig struct {
	Name         string
	Repo         string
	ChartPath    string `mapstructure:",omitempty"`
	DirectCommit bool
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

func (c Config) GetPlatform(platformIndex int) *PlatformConfig {
	if platformIndex < 0 {
		return nil
	}
	return &c.Platforms[platformIndex]
}

func (p *PlatformConfig) GetEnvironment(envIndex int) *EnvironmentConfig {
	return &p.Envs[envIndex]
}
