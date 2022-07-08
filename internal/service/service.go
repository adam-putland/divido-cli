package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"strings"
)

type Service struct {
	gh     *github.GithubClient
	config *models.Config
}

func New(
	gh *github.GithubClient,
	config *models.Config,
) *Service {
	return &Service{
		gh:     gh,
		config: config,
	}
}

func (s *Service) GetChangelog(name string, version1 string, version2 string) (string, error) {
	resp, err := s.gh.GetCommits(s.config.Org, name, version1, version2)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("changelog:\n")
	builder.Grow(len(resp.Commits))
	for _, commit := range resp.Commits {
		_, err := fmt.Fprintf(&builder, "%s\n", *commit.Commit.Message)
		if err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}

func (s *Service) GetServiceLatest(name string) (*models.Release, error) {
	repo, err := s.gh.GetLatestRelease(s.config.Org, name)
	if err != nil {
		return nil, err
	}

	return &models.Release{
		Name:      name,
		Version:   *repo.TagName,
		Changelog: *repo.Body,
		URL:       *repo.HTMLURL,
	}, nil
}

func (s *Service) GetServiceVersions(name string) (models.Releases, error) {
	repository, err := s.gh.GetReleases(s.config.Org, name)
	if err != nil {
		return nil, err
	}
	arr := make([]*models.Release, 0, len(repository))

	for _, repo := range repository {
		arr = append(arr, &models.Release{
			Name:      name,
			Version:   *repo.TagName,
			Changelog: *repo.Body,
			URL:       *repo.HTMLURL,
		})
	}

	return arr, nil
}

func (s Service) GetEnv(ctx context.Context, platIndex, envIndex int) (*models.Environment, error) {

	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return nil, errors.New("could not get platform")
	}

	env := plat.GetEnvironment(envIndex)
	if env == nil {
		return nil, errors.New("could not get env")
	}

	environment := models.Environment{
		EnvironmentConfig: *env,
	}

	//if no ChartPath will load services directly from the env repo
	if env.ChartPath != "" {
		content, err := s.gh.GetContent(ctx, "dividohq",
			env.Repo, env.ChartPath, "master")
		if err != nil {
			return nil, err
		}

		environment.Overrides = content
		return &environment, nil
	}

	hlmVersion, err := s.gh.GetContent(ctx, "dividohq",
		env.Repo, "helm/platform/CURRENT_CHART_VERSION", "master")
	environment.HelmChartVersion = string(hlmVersion)
	if err != nil {
		return nil, err
	}

	if overrides, err := s.gh.GetContent(ctx, "dividohq",
		env.Repo, "helm/platform/versions.yaml", "master"); err == nil {
		environment.Overrides = overrides
	}
	return &environment, nil
}

func (s *Service) LoadEnvServices(ctx context.Context, env *models.Environment, platIndex int) error {

	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return errors.New("could not get platform")
	}

	content, err := s.gh.GetContent(ctx, "dividohq",
		plat.HelmChartRepo, "charts/services/values.yaml", fmt.Sprintf("v%s", strings.TrimSpace(env.HelmChartVersion)))

	if err != nil {
		return err
	}

	env.Services = content
	return nil
}

func (s *Service) UpdateHelmVersion(ctx context.Context, env *models.Environment, githubDetails *models.GitHubCommit, version string) error {
	if env.DirectCommit {
		return s.gh.Commit(ctx, []byte(version), "dividohq", env.Repo, "helm/platform/CURRENT_CHART_VERSION", githubDetails.Branch,
			githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message)
	}

	// todo update via pull request

	return nil
}

func (s *Service) GetHelmVersions(env *models.Environment, platIndex int) (models.Releases, error) {

	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return nil, errors.New("could not get platform")
	}

	repository, err := s.gh.GetReleases(s.config.Org, plat.HelmChartRepo)
	if err != nil {
		return nil, err
	}
	arr := make([]*models.Release, 0, len(repository))

	tagVersion := fmt.Sprintf("v%s", strings.TrimSpace(env.HelmChartVersion))
	for _, repo := range repository {
		if *repo.TagName != tagVersion {
			arr = append(arr, &models.Release{
				Name:      plat.HelmChartRepo,
				Version:   *repo.TagName,
				Changelog: *repo.Body,
				URL:       *repo.HTMLURL,
			})
		}
	}

	return arr, nil
}
