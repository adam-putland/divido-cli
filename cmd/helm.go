package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/sarulabs/di"
	"os"
)

var helmOptions = []string{
	"Info",
	"Bump Service(s)",
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

	fmt.Print(latest)

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
		_, sVersion, err := util.Select("Select last version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		diff, err := s.ComparePlatVersions(ctx, platConfig.HelmChartRepo, fVersion, sVersion)
		if err != nil {
			return fmt.Errorf("comparing versions %w", err)
		}
		fmt.Println(diff)

	case 3:
		return nil
	}

	return HelmOptionsUI(ctx, s, platConfig)

}
