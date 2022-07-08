package models

import "fmt"

var (
	DefaultPreMessage                = "chore(autocommit):"
	DefaultMessageBumpHc             = "Automatic bump hc"
	Default_message_override_service = "Automatic service override"
)

type GitHubCommit struct {
	PullRequestDescription string
	AuthorName             string
	AuthorEmail            string
	Branch                 string
	Message                string
}

func NewGitHubCommit(message string) *GitHubCommit {
	return &GitHubCommit{AuthorName: "dividotech", AuthorEmail: "tech@divido.com", Branch: "master", Message: message}
}

func (g GitHubCommit) String() string {
	return fmt.Sprintf("Author name: %s\n Author email: %s\n Branch: %s\n Message: %s", g.AuthorName, g.AuthorEmail, g.Branch, g.Message)
}
