package internal

import (
	"context"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"time"
)

type GithubClient struct {
	Client      *github.Client
	sourceOwner string
	sourceRepo  string
	chartPath   string
	authorName  string
	authorEmail string
	branch      string
	message     string
}

var MessageType = "blob"
var MODE = "100644"
var BranchHeader = "refs/heads/"

func NewGithubClient(ctx context.Context, token string) *GithubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &GithubClient{
		Client:      github.NewClient(tc),
	}
}

func (c GithubClient) GetChartValues(ctx context.Context) ([]byte, error) {
	chart, _, _, err := c.Client.Repositories.GetContents(ctx, c.sourceOwner, c.sourceRepo, c.chartPath, nil)

	if err != nil {
		return nil, err
	}

	content, err := chart.GetContent()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (c GithubClient) Commit(ctx context.Context, data []byte) error {

	entry := github.TreeEntry{Path: &c.chartPath,
		Type:    &MessageType,
		Content: github.String(string(data)),
		Mode:    &MODE}

	ref, _, err := c.Client.Git.GetRef(ctx, c.sourceOwner, c.sourceRepo, BranchHeader+c.branch)
	if err != nil {
		return err
	}

	tree, _, err := c.Client.Git.CreateTree(ctx, c.sourceOwner, c.sourceRepo, *ref.Object.SHA, []*github.TreeEntry{&entry})
	if err != nil {
		return err
	}

	parent, _, err := c.Client.Repositories.GetCommit(ctx, c.sourceOwner, c.sourceRepo, *ref.Object.SHA, nil)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &c.authorName, Email: &c.authorEmail}
	commit := &github.Commit{Author: author, Message: &c.message, Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := c.Client.Git.CreateCommit(ctx, c.sourceOwner, c.sourceRepo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = c.Client.Git.UpdateRef(ctx, c.sourceOwner, c.sourceRepo, ref, true)
	return err
}
