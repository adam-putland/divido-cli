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

	builder := strings.Builder{}
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

func (s *Service) GetServiceVersion(name string) (*models.Release, error) {
	return nil, nil
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

func (s Service) GetEnv(ctx context.Context, projectIndex, envIndex int) (*models.Environment, error) {

	proj := s.config.GetPlatform(projectIndex)
	if proj == nil {
		return nil, errors.New("could not get project")
	}

	env := proj.GetEnvironment(envIndex)
	if env == nil {
		return nil, errors.New("could not get env")
	}

	// if ChartPath will load services directly from the env repo
	//if env.ChartPath != "" {
	//	content, err := s.gh.GetContent(ctx, "dividohq",
	//		env.Repo, env.ChartPath, "master")
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//hlmVersion, err := s.gh.GetContent(ctx, "dividohq",
	//	env.Repo, "helm/platform/CURRENT_CHART_VERSION", "master")
	//if err != nil {
	//	return nil, err
	//}
	//
	//return s.gh.GetContent(ctx, "dividohq",
	//	proj.HelmChartRepo, "charts/services/values.yaml", fmt.Sprintf("v%s", strings.TrimSpace(string(hlmVersion))))
	return nil, nil
}
