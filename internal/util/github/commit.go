package github

import (
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
)

type Commit struct {
	PullRequestDescription string
	PullRequestTitle       string
	AuthorName             string
	AuthorEmail            string
	Branch                 string
	Message                string
	Org                    string
}

func NewGitHubCommit(config *models.GithubConfig) *Commit {
	return &Commit{
		AuthorName:  config.AuthorName,
		AuthorEmail: config.AuthorEmail,
		Branch:      config.MainBranch,
		Org:         config.Org,
	}
}

func WithBumpHC(config *models.GithubConfig, version string) *Commit {

	message := fmt.Sprintf("%s: %s %s", config.PreCommitMessage, config.CommitMessageBumpHc, version)
	commit := NewGitHubCommit(config)
	commit.Branch = fmt.Sprintf("chore/bump-hc-%s", version)
	commit.Message = message
	commit.PullRequestTitle = message
	commit.PullRequestDescription = message
	return commit
}

func WithBumpServices(config *models.GithubConfig) *Commit {
	commit := NewGitHubCommit(config)
	commit.Message = fmt.Sprintf("%s: %s", config.PreCommitMessage, config.CommitMessageBumpService)
	return commit
}

func (c Commit) String() string {
	return fmt.Sprintf(" Org: %s\n Author name: %s\n Author email: %s\n Branch: %s\n Commit message: %s", c.Org, c.AuthorName, c.AuthorEmail, c.Branch, c.Message)
}

func (c Commit) PullRequestInfo() string {
	return fmt.Sprintf("\n Pull request title: %s\n Pull request description: %s", c.PullRequestTitle, c.PullRequestDescription)
}
