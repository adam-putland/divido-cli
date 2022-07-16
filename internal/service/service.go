package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strings"
)

var (
	_defaultChartVersionFilePath  = "helm/platform/CURRENT_CHART_VERSION"
	_defaultHelmOverridesFilePath = "helm/platform/versions.yaml"
	_defaultChatServicesFilePath  = "charts/services/values.yaml"
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

func (s *Service) GetChangelog(ctx context.Context, name string, version1 string, version2 string) (string, error) {
	resp, err := s.gh.GetCommits(ctx, s.config.Github.Org, name, version1, version2)
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

func (s *Service) GetLatest(ctx context.Context, name string) (*models.Release, error) {
	repo, err := s.gh.GetLatestRelease(ctx, s.config.Github.Org, name)
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

func (s *Service) GetRepoReleases(ctx context.Context, name string) (models.Releases, error) {
	repository, err := s.gh.GetReleases(ctx, s.config.Github.Org, name)
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

	// if no ChartPath will load services directly from the env repo
	if env.ChartPath != "" {
		content, err := s.gh.GetContent(ctx, s.config.Github.Org,
			env.Repo, env.ChartPath, s.config.Github.MainBranch)
		if err != nil {
			return nil, err
		}

		environment.Overrides = content
		return &environment, nil
	}

	hlmVersion, err := s.gh.GetContent(ctx, s.config.Github.Org,
		env.Repo, _defaultChartVersionFilePath, s.config.Github.MainBranch)
	environment.HelmChartVersion = string(hlmVersion)
	if err != nil {
		return nil, err
	}

	if overrides, err := s.gh.GetContent(ctx, s.config.Github.Org,
		env.Repo, _defaultHelmOverridesFilePath, s.config.Github.MainBranch); err == nil {
		environment.Overrides = overrides
	}
	return &environment, nil
}

func (s *Service) LoadEnvServices(ctx context.Context, env *models.Environment, platIndex int) error {

	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return errors.New("could not get platform")
	}

	content, err := s.gh.GetContent(ctx, s.config.Github.Org,
		plat.HelmChartRepo, _defaultChatServicesFilePath, env.GetHCVersion())

	if err != nil {
		return err
	}

	env.Services = content
	return nil
}

func (s *Service) UpdateHelmVersion(ctx context.Context, env *models.Environment, githubDetails *github.Commit, version string) error {

	version = strings.Trim(version, "v")
	if env.DirectCommit {
		err := s.gh.Commit(ctx, []byte(version), s.config.Github.Org, env.Repo, _defaultChartVersionFilePath, githubDetails.Branch,
			githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message)
		if err != nil {
			return err
		}
	}

	err := s.gh.CreatePullRequest(ctx, []byte(version), s.config.Github.Org, env.Repo, _defaultChartVersionFilePath, githubDetails.Branch,
		s.config.Github.MainBranch, githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message, githubDetails.PullRequestTitle, githubDetails.PullRequestDescription)
	if err != nil {
		return err
	}

	env.HelmChartVersion = version
	return nil
}

func (s *Service) GetHelmVersions(ctx context.Context, env *models.Environment, platIndex int) (models.Releases, error) {

	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return nil, errors.New("could not get platform")
	}

	repository, err := s.gh.GetReleases(ctx, s.config.Github.Org, plat.HelmChartRepo)
	if err != nil {
		return nil, err
	}
	arr := make([]*models.Release, 0, len(repository))

	tagVersion := env.GetHCVersion()
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

func (s *Service) GetPlat(ctx context.Context, platRepo string) (*models.Platform, error) {

	latest, err := s.GetLatest(ctx, platRepo)
	if err != nil {
		return nil, err
	}

	plat := models.Platform{Release: *latest}

	plat.Services, err = s.gh.GetContent(ctx, s.config.Github.Org, platRepo, _defaultChatServicesFilePath, latest.Version)
	if err != nil {
		return nil, err
	}

	return &plat, nil
}

func (s *Service) ComparePlatVersions(ctx context.Context, repo string, version string, version2 string) (*models.Comparer, error) {

	resultsChan := make(chan string)

	changelog, err := s.GetChangelog(ctx, repo, version, version2)
	if err != nil {
		return nil, err
	}

	go func() {
		content, err := s.gh.GetContent(ctx, s.config.Github.Org, repo, _defaultChatServicesFilePath, version)
		if err != nil {
			resultsChan <- ""
		}
		resultsChan <- string(content)
	}()

	go func() {
		content, err := s.gh.GetContent(ctx, s.config.Github.Org, repo, _defaultChatServicesFilePath, version2)
		if err != nil {
			resultsChan <- ""
		}
		resultsChan <- string(content)
	}()

	comparer := models.Comparer{Changelog: changelog}
	var results []string
	for {
		result := <-resultsChan
		results = append(results, result)

		// if we've reached the expected amount of urls then stop
		if len(results) == 2 {
			break
		}
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(results[0], results[1], false)
	comparer.Diff = dmp.DiffPrettyText(diffs)

	return &comparer, nil
}
