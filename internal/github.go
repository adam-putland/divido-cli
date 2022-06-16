package internal

import (
	"context"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"time"
)

type GithubClient struct {
	Client *github.Client
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
		Client: github.NewClient(tc),
	}
}

func (c GithubClient) GetChartValues(ctx context.Context, sourceOwner string, sourceRepo string, chartPath string) ([]byte, error) {

	if chartPath == "" {
		chartPath = "config/values.yaml"
	}
	chart, _, _, err := c.Client.Repositories.GetContents(ctx, sourceOwner, sourceRepo, chartPath, nil)

	if err != nil {
		return nil, err
	}

	content, err := chart.GetContent()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (c GithubClient) Commit(ctx context.Context, data []byte, sourceOwner string, sourceRepo string, chartPath string,
	branch string, authorName string, authorEmail string, message string) error {

	entry := github.TreeEntry{Path: &chartPath,
		Type:    &MessageType,
		Content: github.String(string(data)),
		Mode:    &MODE}

	ref, _, err := c.Client.Git.GetRef(ctx, sourceOwner, sourceRepo, BranchHeader+branch)
	if err != nil {
		return err
	}

	tree, _, err := c.Client.Git.CreateTree(ctx, sourceOwner, sourceRepo, *ref.Object.SHA, []*github.TreeEntry{&entry})
	if err != nil {
		return err
	}

	parent, _, err := c.Client.Repositories.GetCommit(ctx, sourceOwner, sourceRepo, *ref.Object.SHA, nil)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &authorName, Email: &authorEmail}
	commit := &github.Commit{Author: author, Message: &message, Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := c.Client.Git.CreateCommit(ctx, sourceOwner, sourceRepo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = c.Client.Git.UpdateRef(ctx, sourceOwner, sourceRepo, ref, true)
	return err
}
