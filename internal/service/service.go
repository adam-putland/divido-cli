package service

import (
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"strings"
)

type Service struct {
	gh               *github.GithubClient
	config 			*models.Config
}

func New(
	gh *github.GithubClient,
	config *models.Config,
) *Service {
	return &Service{
		gh:               gh,
		config: config,
	}
}

func (s *Service) GetChangelog(name string, version1 string, version2 string) (string, error) {
	resp, err := s.gh.GetCommits(s.config.Org, name, version1, version2)
	if err != nil {
		return "", err
	}

	builder := strings.Builder{}
	for _, commit := range resp.Commits {
		builder.WriteString(*commit.Commit.Message)
	}
	return builder.String(), nil
}

func (s* Service) GetServiceLatest(name string) (*models.Service, error) {
	repo, err := s.gh.GetLatestRelease(s.config.Org, name)
	if err != nil {
		return nil, err
	}

	return &models.Service{
		Name: name,
		Version: *repo.TagName,
		Body: *repo.Body,
		URL: *repo.HTMLURL,
	}, nil
}

func (s* Service) GetServiceVersion(name string) (*models.Service, error) {
	return nil, nil
}

func (s *Service) GetServiceVersions(name string) ([]models.Service, error) {
	repository, err := s.gh.GetReleases(s.config.Org, name)
	if err != nil {
		return nil, err
	}
	arr := make([]models.Service, 0, len(repository))

	for _, repo := range repository {
		service := models.Service{
			Name: name,
			Version: *repo.TagName,
			Body: *repo.Body,
			URL: *repo.HTMLURL,
		}
		arr = append(arr, service)
	}

	return arr, nil
}

//services -> enter service (with search) -> service
//-> versions
//-> changelog
//->