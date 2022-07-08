package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/ui"
	"github.com/sarulabs/di"
	"os"
	"strings"
)

var envOptions = []string{
	"Services",
	"Bump helm version",
	"Back",
}

func EnvUI(ctx context.Context, app di.Container) error {
	s := app.Get("service").(*service.Service)
	config := app.Get("config").(*models.Config)
	platIndex, _, err := ui.Select("Select platform", config.ListPlatform())
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	envI, _, err := ui.Select("Select env", config.ListEnvironments(platIndex))
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	env, err := s.GetEnv(ctx, platIndex, envI)
	if err != nil {
		return err
	}

	fmt.Println(env.Info())

	return EnvOptionsUI(ctx, s, env, platIndex)
}

func EnvOptionsUI(ctx context.Context, s *service.Service, env *models.Environment, platIndex int) error {

	option, _, err := ui.Select("Choose option", envOptions)
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

		releases, err := s.GetHelmVersions(env, platIndex)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()
		_, fVersion, err := ui.Select("Select version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		version := strings.Trim(fVersion, "v")

		message := fmt.Sprintf("%s: %s %s", models.DefaultPreMessage, models.DefaultMessageBumpHc, fVersion)
		githubDetails := models.NewGitHubCommit(message)
		err = BumpHelmUI(ctx, s, env, githubDetails, version)
		if err != nil {
			fmt.Println(err)
		}

	case 2:
		return nil
	}

	return EnvOptionsUI(ctx, s, env, platIndex)
}

func BumpHelmUI(ctx context.Context, s *service.Service, env *models.Environment, githubDetails *models.GitHubCommit, version string) error {
	fmt.Printf("Github Details \n%s", githubDetails)

	options := []string{
		"Change Author Name",
		"Change Author Email",
		"Change Commit Message",
		"Change Branch",
		"Continue",
		"Back",
	}
	if !env.DirectCommit {
		fmt.Printf("Pull Request Description %s", githubDetails.PullRequestDescription)
		options = append(options, "Change pull request description")
	}

	githubC, _, err := ui.Select("Choose option", options)
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	switch githubC {
	case 0:
		githubDetails.AuthorName, err = ui.PromptWithDefault("Enter Author Name", githubDetails.AuthorName)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}

	case 1:
		githubDetails.AuthorName, err = ui.PromptWithDefault("Enter Author Name", githubDetails.AuthorName)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	case 2:
		githubDetails.AuthorName, err = ui.PromptWithDefault("Enter Author Name", githubDetails.AuthorName)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	case 3:
		githubDetails.AuthorName, err = ui.PromptWithDefault("Enter Author Name", githubDetails.AuthorName)
		if err != nil {
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	case 4:
		err = s.UpdateHelmVersion(ctx, env, githubDetails, version)
		if err != nil {
			return fmt.Errorf("loading environment services %w", err)
		}
		fmt.Printf("Helm updated to version %s", version)
		return nil
	}
	return BumpHelmUI(ctx, s, env, githubDetails, version)
}
