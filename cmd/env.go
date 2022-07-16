package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/sarulabs/di"
	"os"
)

var envOptions = []string{
	"Services",
	"Bump helm version",
	"Back",
}

func EnvUI(ctx context.Context, app di.Container) error {
	s := app.Get("service").(*service.Service)
	config := app.Get("config").(*models.Config)
	platIndex, _, err := util.Select("Select platform", config.ListPlatform())
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	envI, _, err := util.Select("Select env", config.ListEnvironments(platIndex))
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	env, err := s.GetEnv(ctx, platIndex, envI)
	if err != nil {
		return err
	}

	fmt.Println(env)

	return EnvOptionsUI(ctx, s, env, config, platIndex)
}

func EnvOptionsUI(ctx context.Context, s *service.Service, env *models.Environment, config *models.Config, platIndex int) error {

	option, _, err := util.Select("Choose option", envOptions)
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	switch option {
	case 0:
		err := s.LoadEnvServices(ctx, env, platIndex)
		if err != nil {
			return fmt.Errorf("loading environment services %w", err)
		}
		fmt.Println(string(env.Services))
	case 1:
		fmt.Printf("Current Version: %s", env.HelmChartVersion)

		releases, err := s.GetHelmVersions(ctx, env, platIndex)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()
		_, fVersion, err := util.Select("Select version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		githubDetails := github.WithBumpHC(&config.Github, fVersion)
		err = BumpHelmUI(ctx, s, env, githubDetails, fVersion)
		if err != nil {
			fmt.Println(err)
		}

	case 2:
		return nil
	}

	return EnvOptionsUI(ctx, s, env, config, platIndex)
}

func BumpHelmUI(ctx context.Context, s *service.Service, env *models.Environment, gd *github.Commit, version string) error {
	fmt.Printf("Github Details \n%s", gd)

	options := []string{
		"Change Author Name",
		"Change Author Email",
		"Change Commit Message",
		"Change Branch",
		"Continue",
		"Back",
	}

	if !env.DirectCommit {
		fmt.Print(gd.PullRequestInfo())
		options = append(options, []string{"Change pull request title", "Change pull request description"}...)
	}

	githubC, _, err := util.Select("Choose option", options)
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	switch githubC {
	case 0:
		gd.AuthorName, err = util.PromptWithDefault("Enter Author Name", gd.AuthorName)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}

	case 1:
		gd.AuthorName, err = util.PromptWithDefault("Enter Author Email", gd.AuthorEmail)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	case 2:
		gd.AuthorName, err = util.PromptWithDefault("Enter Commit Message", gd.Message)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	case 3:
		gd.AuthorName, err = util.PromptWithDefault("Enter Branch", gd.Branch)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	case 4:

		err = s.UpdateHelmVersion(ctx, env, gd, version)
		if err != nil {
			return fmt.Errorf("loading environment services %w", err)
		}
		fmt.Printf("Env: %s Helm updated to version %s", env.Name, version)
		return nil
	case 5:
		return nil
	case 6:
		gd.PullRequestTitle, err = util.PromptWithDefault("Enter Pull request title", gd.PullRequestTitle)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}

	case 7:
		gd.PullRequestDescription, err = util.PromptWithDefault("Enter Pull request description", gd.PullRequestDescription)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}

	}
	return BumpHelmUI(ctx, s, env, gd, version)
}
