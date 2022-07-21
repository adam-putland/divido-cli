package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/gobeam/stringy"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	_defaultChartVersionFilePath  = "helm/platform/CURRENT_CHART_VERSION"
	_defaultHelmOverridesFilePath = "helm/platform/versions.yaml"
	_defaultChatServicesFilePath  = "charts/services/values.yaml"
	_defaultReleasesPath          = "./releases"
	_defaultReleaseFileName       = "JIRA_TICKET_TEXT.txt"
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

func (s Service) GetConfig() *models.Config {
	return s.config
}

func (s *Service) GetChangelog(ctx context.Context, name string, release1, release2 *models.Release) (string, error) {

	version1 := release1.Version
	version2 := release2.Version

	if release1.Date.After(release2.Date) {
		version1 = release2.Version
		version2 = release1.Version
	}

	resp, err := s.gh.GetChangelog(ctx, s.config.Github.Org, name, version1, version2)
	if err != nil {
		return "", err
	}

	return resp.Body, nil
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
			Date:      repo.PublishedAt.Time,
		})
	}

	return arr, nil
}

func (s Service) GetEnv(ctx context.Context, platIndex, envIndex int) (*models.Environment, error) {

	platCfg := s.config.GetPlatform(platIndex)
	if platCfg == nil {
		return nil, util.ErrMissingPlat
	}

	envCfg := platCfg.GetEnvironment(envIndex)
	if envCfg == nil {
		return nil, errors.New("could not get env")
	}

	env := models.Environment{
		EnvironmentConfig: *envCfg,
	}

	if !env.OnlyOverrides {
		hlmVersion, err := s.gh.GetContent(ctx, s.config.Github.Org,
			env.Repo, _defaultChartVersionFilePath, s.config.Github.MainBranch)
		env.HelmChartVersion = string(hlmVersion)
		if err != nil {
			return nil, err
		}
	}

	return &env, nil
}

func (s *Service) LoadEnvServices(ctx context.Context, env *models.Environment, platIndex int) error {

	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return util.ErrMissingPlat
	}

	if !env.OnlyOverrides {
		content, err := s.gh.GetContent(ctx, s.config.Github.Org,
			plat.HelmChartRepo, _defaultChatServicesFilePath, env.GetHCVersion())
		if err != nil {
			return err
		}

		services, err := NewParser(content).Load()
		if err != nil {
			return err
		}

		env.Services = services
	}

	// if no ChartPath will load services directly from the env repo
	if env.ChartPath != "" {
		content, err := s.gh.GetContent(ctx, s.config.Github.Org,
			env.Repo, env.ChartPath, s.config.Github.MainBranch)
		if err != nil {
			return err
		}

		overrides, err := NewParser(content).Load()
		if err != nil {
			return err
		}
		env.Overrides = overrides
	}

	if content, err := s.gh.GetContent(ctx, s.config.Github.Org,
		env.Repo, _defaultHelmOverridesFilePath, s.config.Github.MainBranch); err == nil {

		overrides, err := NewParser(content).Load()
		if err != nil {
			return err
		}
		env.Overrides = overrides
	}
	return nil
}

func (s *Service) UpdateHelmVersion(ctx context.Context, env *models.Environment, githubDetails *github.Commit, version string) error {

	version = strings.Trim(version, "v")
	data := []byte(version)
	if env.DirectCommit {
		err := s.gh.Commit(ctx, data, s.config.Github.Org, env.Repo, _defaultChartVersionFilePath, githubDetails.Branch,
			githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message)
		if err != nil {
			return err
		}
	} else {
		err := s.gh.CreatePullRequest(ctx, data, s.config.Github.Org, env.Repo, _defaultChartVersionFilePath, githubDetails.Branch,
			s.config.Github.MainBranch, githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message, githubDetails.PullRequestTitle, githubDetails.PullRequestDescription)
		if err != nil {
			return err
		}
	}

	env.HelmChartVersion = version
	return nil
}

func (s *Service) GetHelmVersions(ctx context.Context, env *models.Environment, platIndex int) (models.Releases, error) {
	plat := s.config.GetPlatform(platIndex)
	if plat == nil {
		return nil, util.ErrMissingPlat
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

	plat := models.Platform{Release: latest}

	content, err := s.gh.GetContent(ctx, s.config.Github.Org, platRepo, _defaultChatServicesFilePath, latest.Version)
	if err != nil {
		return nil, err
	}

	parser := NewParser(content)

	plat.Services, err = parser.Load()
	if err != nil {
		return nil, err
	}

	return &plat, nil
}

func (s *Service) UpdateServicesVersions(ctx context.Context, platCfg *models.PlatformConfig, githubDetails *github.Commit, servicesUpdated []*models.ServiceUpdated) error {

	latest, err := s.GetLatest(ctx, platCfg.HelmChartRepo)
	if err != nil {
		return err
	}

	content, err := s.gh.GetContent(ctx, s.config.Github.Org, platCfg.HelmChartRepo, _defaultChatServicesFilePath, latest.Version)
	if err != nil {
		return err
	}

	parser := NewParser(content)

	_, err = parser.Load()
	if err != nil {
		return err
	}

	// some validations could take place here from the parser.Load()

	servicesToReplace := make(models.Services, len(servicesUpdated))
	for _, updated := range servicesUpdated {
		newService := *updated.Service
		newService.Version = updated.NewVersion
		servicesToReplace[updated.Service.Name] = &newService

	}

	err = parser.Replace(servicesToReplace)
	if err != nil {
		return err
	}

	content, err = parser.GetContent()
	if err != nil {
		return err
	}

	if platCfg.DirectCommit {
		return s.gh.Commit(ctx, content, s.config.Github.Org, platCfg.HelmChartRepo, _defaultChatServicesFilePath, githubDetails.Branch,
			githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message)
	}

	return s.gh.CreatePullRequest(ctx, content, s.config.Github.Org, platCfg.HelmChartRepo, _defaultChatServicesFilePath, githubDetails.Branch,
		s.config.Github.MainBranch, githubDetails.AuthorName, githubDetails.AuthorEmail, githubDetails.Message, githubDetails.PullRequestTitle, githubDetails.PullRequestDescription)
}

func (s *Service) ComparePlatReleasesByVersion(ctx context.Context, platCfg *models.PlatformConfig, releases models.Releases, version string, version2 string) (*models.Comparer, error) {
	resultsChan := make(chan *models.Platform)

	for _, v := range []string{version, version2} {
		go func(version string) {
			content, err := s.gh.GetContent(ctx, s.config.Github.Org, platCfg.HelmChartRepo, _defaultChatServicesFilePath, version)
			if err != nil {
				fmt.Println(err)
				resultsChan <- nil
			}

			parser := NewParser(content)

			services, err := parser.Load()
			if err != nil {
				fmt.Println(err)
				resultsChan <- nil
			}

			resultsChan <- &models.Platform{
				Release:  releases.GetReleaseByVersion(version),
				Services: services,
			}
		}(v)
	}

	var results []*models.Platform
	for {
		result := <-resultsChan

		if result == nil {
			return nil, errors.New("could not get content for specified versions")
		}

		results = append(results, result)
		// if we've reached the expected amount of results then stop
		if len(results) == 2 {
			break
		}
	}

	return models.Compare(results[0], results[1]), nil
}

func (s Service) GetChangelogsFromDiff(ctx context.Context, diff *models.Comparer) (map[string]string, error) {

	changelogs := make(map[string]string, len(diff.Changed)+len(diff.Insert))
	for serviceName, changed := range diff.Changed {

		repoName, multi := s.ServiceNameToKebabCase(serviceName)

		if _, ok := changelogs[repoName]; !ok {

			version1 := changed.Service.Version
			version2 := changed.NewVersion
			if multi {
				version1 = fmt.Sprintf("%s-%s", stringy.New(serviceName).KebabCase().ToLower(), version1)
				version2 = fmt.Sprintf("%s-%s", stringy.New(serviceName).KebabCase().ToLower(), version2)
			}
			resp, err := s.gh.GetChangelog(ctx, s.config.Github.Org, repoName, version1, version2)
			if err != nil {
				return nil, err
			}

			changelogs[repoName] = resp.Body
		}

	}

	for serviceName, service := range diff.Insert {

		repoName, _ := s.ServiceNameToKebabCase(serviceName)
		releases, err := s.GetRepoReleases(ctx, repoName)
		if err != nil {
			return nil, err
		}

		var builder strings.Builder
		release := releases.GetReleaseByVersion(service.Version)
		builder.WriteString(release.Changelog)

		for _, r := range releases {
			if r.Date.Before(release.Date) {
				builder.WriteString(r.Changelog)
			}
		}

		changelogs[repoName] = builder.String()

	}

	return changelogs, nil
}

func (s Service) ExportRelease(ctx context.Context, diff *models.Comparer) error {

	if _, err := os.Stat(_defaultReleasesPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(_defaultReleasesPath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	releaseFilePath := fmt.Sprintf("%s/release-%s-%s-%s", _defaultReleasesPath, diff.InitialVersion, diff.FinalVersion, time.Now().UTC().Format("02-01-06"))
	if _, err := os.Stat(releaseFilePath); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(releaseFilePath, os.ModePerm); err != nil {
			log.Println(err)
		}
	}

	f, err := os.OpenFile(filepath.Join(releaseFilePath, _defaultReleaseFileName), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	diff.DisableColor = true
	if _, err := f.WriteString(diff.String()); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	changelogs, err := s.GetChangelogsFromDiff(ctx, diff)
	if err != nil {
		return err
	}

	for service, changelog := range changelogs {

		filename := fmt.Sprintf("%s_changelog.txt", service)
		f, err := os.OpenFile(filepath.Join(releaseFilePath, filename), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.WriteString(changelog); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil

}

func (s Service) GetAvailableServiceReleases(ctx context.Context, service *models.Service) (models.Releases, error) {

	repoName, multi := s.ServiceNameToKebabCase(service.Name)

	version := service.Version
	if multi {
		version = fmt.Sprintf("%s-%s", stringy.New(service.Name).KebabCase().Get(), version)
	}

	release, err := s.gh.GetRelease(ctx, s.config.Github.Org, repoName, version)
	if err != nil {
		return nil, err
	}

	repoReleases, err := s.gh.GetReleases(ctx, s.config.Github.Org, repoName)
	if err != nil {
		return nil, err
	}
	releases := make([]*models.Release, 0, len(repoReleases))

	for _, repo := range repoReleases {

		if release.CreatedAt.After(repo.CreatedAt.Time) || release.CreatedAt.Equal(*repo.CreatedAt) {
			continue
		}
		releases = append(releases, &models.Release{
			Name:      repoName,
			Version:   *repo.TagName,
			Changelog: *repo.Body,
			URL:       *repo.HTMLURL,
			Date:      repo.CreatedAt.Time,
		})
	}

	return releases, err
}

func (s Service) ServiceNameToKebabCase(serviceName string) (string, bool) {
	repoName := stringy.New(serviceName).KebabCase().Get()
	multiTag := false
	for regex, mapping := range s.config.ServicesMapping {
		if matched, _ := regexp.MatchString(regex, serviceName); matched {
			repoName = mapping.Repo
			multiTag = mapping.MultiTag
		}
	}
	return strings.ToLower(repoName), multiTag
}
