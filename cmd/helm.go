package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/manifoldco/promptui"
	"github.com/sarulabs/di"
	"os"
)

var helmOptions = []string{
	"Info",
	"Bump Service(s) (In development)",
	"Compare Versions",
	"Back",
}

func HelmUI(ctx context.Context, app di.Container) error {

	s := app.Get("service").(*service.Service)
	config := app.Get("config").(*models.Config)
	platIndex, _, err := util.Select("Select platform", config.ListPlatform())
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	platConfig := config.GetPlatform(platIndex)

	return HelmOptionsUI(ctx, s, platConfig)

}

func HelmOptionsUI(ctx context.Context, s *service.Service, platConfig *models.PlatformConfig) error {

	latest, err := s.GetLatest(ctx, platConfig.HelmChartRepo)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Latest Release: \n%s", latest)

	option, _, err := util.Select("Choose option", helmOptions)
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	switch option {
	case 0:

		plat, err := s.GetPlat(ctx, platConfig.HelmChartRepo)
		if err != nil {
			return fmt.Errorf("getting services in hlm %w", err)
		}
		fmt.Println(plat)

	case 1:
		// TODO

	case 2:
		releases, err := s.GetRepoReleases(ctx, platConfig.HelmChartRepo)
		if err != nil {
			return fmt.Errorf("getting platform versions %w", err)
		}

		versions := releases.Versions()
		fi, fVersion, err := util.Select("Select first version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		versions.Remove(fi)
		_, lVersion, err := util.Select("Select last version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		diff, err := s.ComparePlatReleasesByVersion(ctx, platConfig, releases, fVersion, lVersion)
		if err != nil {
			return fmt.Errorf("comparing versions %w", err)
		}
		fmt.Println(diff)

		err = VersionsUI(ctx, s, diff)
		if err != nil {
			return err
		}

	case 3:
		return nil
	}

	return HelmOptionsUI(ctx, s, platConfig)
}

func VersionsUI(ctx context.Context, s *service.Service, diff *models.Comparer) error {

	options := []string{
		"Show Changelogs",
		"Export Release",
		"Create Release Ticket (In development)",
		"Back",
	}

	option, _, err := util.Select("Choose option", options)
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	switch option {
	case 0:
		changelogs, err := s.GetChangelogsFromDiff(ctx, diff)
		if err != nil {
			return err
		}

		for key, changelog := range changelogs {
			fmt.Printf("\nService: %s\n", key)
			fmt.Println(changelog)
		}
	case 1:
		err := s.ExportRelease(ctx, diff)
		if err != nil {
			fmt.Println(promptui.IconBad + " Release not exported")
			return err
		}
		fmt.Println(promptui.IconGood + " Release exported")

	case 2:

	case 3:
		return nil
	}

	return VersionsUI(ctx, s, diff)
}
