package internal

type Config struct {
	Projects []Project
}

type Project struct {
	Name      string
	HelmChart string `mapstructure:"hlm"`
	Envs      []Environment
}

type Environment struct {
	Name      string
	Repo      string
	ChartPath string `mapstructure:",omitempty"`
}

func (c Config) ListProject() []string {
	projects := make([]string, 0, len(c.Projects))
	for _, project := range c.Projects {
		projects = append(projects, project.Name)
	}
	return projects
}

func (c Config) ListEnvironments(projectIndex int) []string {
	p := c.GetProject(projectIndex)

	if p == nil {
		return nil
	}

	environments := make([]string, 0, len(p.Envs))

	for _, env := range p.Envs {
		environments = append(environments, env.Name)
	}

	return environments
}

func (c Config) GetProject(projectIndex int) *Project {
	if projectIndex < 0 {
		return nil
	}
	return &c.Projects[projectIndex]
}

func (c Config) GetEnvironment(projectIndex, envIndex int) *Environment {
	p := c.GetProject(projectIndex)
	if p == nil {
		return nil
	}
	return &p.Envs[envIndex]
}
