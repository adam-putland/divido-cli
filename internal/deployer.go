package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Deployer struct {
	gh     *GithubClient
	parser *Parser
	config *Config
}

func NewDeployer(ctx context.Context, config *Config, token string) *Deployer {

	return &Deployer{gh: NewGithubClient(ctx, token), config: config}
}

func (d Deployer) GetEnvServices(ctx context.Context, projectIndex, envIndex int) ([]byte, error) {

	proj := d.config.GetProject(projectIndex)
	if proj == nil {
		return nil, errors.New("could not get project")
	}

	env := proj.GetEnvironment(envIndex)
	if env == nil {
		return nil, errors.New("could not get env")
	}

	// if ChartPath will load services directly from the env repo
	if env.ChartPath != "" {
		return d.gh.GetContent(ctx, "dividohq",
			env.Repo, env.ChartPath, "master")
	}

	hlmVersion, err := d.gh.GetContent(ctx, "dividohq",
		env.Repo, "helm/platform/CURRENT_CHART_VERSION", "master")
	if err != nil {
		return nil, err
	}

	return d.gh.GetContent(ctx, "dividohq",
		proj.HelmChart, "charts/services/values.yaml", fmt.Sprintf("v%s", strings.TrimSpace(string(hlmVersion))))

}
