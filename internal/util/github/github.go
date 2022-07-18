package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"time"
)

type GithubClient struct {
	Client *github.Client
}

var (
	_messageType  = "blob"
	_mode         = "100644"
	_branchHeader = "refs/heads/"
)

func NewGithubClient(ctx context.Context, token string) *GithubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &GithubClient{
		Client: github.NewClient(tc),
	}
}

func (c GithubClient) GetContent(ctx context.Context, sourceOwner, sourceRepo, filePath, ref string) ([]byte, error) {

	contentFile, _, _, err := c.Client.Repositories.GetContents(ctx, sourceOwner, sourceRepo, filePath, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	if err != nil {
		return nil, err
	}

	content, err := contentFile.GetContent()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (c GithubClient) Commit(ctx context.Context, data []byte, sourceOwner string, sourceRepo string, filePath string,
	branch string, authorName string, authorEmail string, message string) error {

	entry := github.TreeEntry{Path: &filePath,
		Type:    &_messageType,
		Content: github.String(string(data)),
		Mode:    &_mode}

	ref, _, err := c.Client.Git.GetRef(ctx, sourceOwner, sourceRepo, _branchHeader+branch)
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

func (c *GithubClient) GetChangelog(ctx context.Context, org string, repo string, base string, head string) (*github.RepositoryReleaseNotes, error) {
	res, _, err := c.Client.Repositories.GenerateReleaseNotes(ctx, org, repo, &github.GenerateNotesOptions{
		TagName:         head,
		PreviousTagName: github.String(base),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *GithubClient) GetReleases(ctx context.Context, org string, repo string) ([]*github.RepositoryRelease, error) {
	res, _, err := c.Client.Repositories.ListReleases(ctx, org, repo, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *GithubClient) GetLatestRelease(ctx context.Context, org string, repo string) (*github.RepositoryRelease, error) {
	res, _, err := c.Client.Repositories.GetLatestRelease(ctx, org, repo)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c GithubClient) CreatePullRequest(ctx context.Context, data []byte,
	sourceOwner, sourceRepo, filePath, commitBranch, baseBranch, authorName, authorEmail, message, prTitle, prDescription string) error {

	if ref, _, err := c.Client.Git.GetRef(ctx, sourceOwner, sourceRepo, _branchHeader+commitBranch); err == nil || ref != nil {
		return fmt.Errorf("branch %s already exists", commitBranch)
	}

	// create branch if ref not found
	if commitBranch == baseBranch {
		return fmt.Errorf("the commit branch: %s cannot be the same as the base branch: %s", commitBranch, baseBranch)
	}

	if baseBranch == "" {
		return errors.New("the base branch should not be set to an empty string")
	}

	baseRef, _, err := c.Client.Git.GetRef(ctx, sourceOwner, sourceRepo, _branchHeader+baseBranch)
	if err != nil {
		return err
	}

	newRef := &github.Reference{Ref: github.String(_branchHeader + commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err := c.Client.Git.CreateRef(ctx, sourceOwner, sourceRepo, newRef)
	if err != nil {
		return err
	}

	entry := github.TreeEntry{Path: &filePath,
		Type:    &_messageType,
		Content: github.String(string(data)),
		Mode:    &_mode}

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

	// Attach the commit to the branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = c.Client.Git.UpdateRef(ctx, sourceOwner, sourceRepo, ref, true)
	if err != nil {
		return err
	}

	newPR := &github.NewPullRequest{
		Title:               &prTitle,
		Head:                &commitBranch,
		Base:                &baseBranch,
		Body:                &prDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := c.Client.PullRequests.Create(ctx, sourceOwner, sourceRepo, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("pr created at: %s\n", pr.GetHTMLURL())
	return nil
}
