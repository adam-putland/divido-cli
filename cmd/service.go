package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/ui"
	"github.com/sarulabs/di"
	"os"
)

var serviceOptions = []string{
	"Versions",
	"Generate Changelog",
	"Back",
}

func ServiceUI(ctx context.Context, app di.Container) error {
	s := app.Get("service").(*service.Service)
	serviceName, err := ui.Prompt("Enter service")
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	serv, err := s.GetServiceLatest(serviceName)
	if err != nil {
		return fmt.Errorf("getting service %w", err)
	}
	fmt.Println(serv.Info())

	return ServiceOptionsUI(ctx, s, serviceName)
}

func ServiceOptionsUI(ctx context.Context, s *service.Service, serviceName string) error {
	option, _, err := ui.Select("Choose option", serviceOptions)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	switch option {
	case 0:
		versions, err := s.GetServiceVersions(serviceName)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}
		fmt.Print(versions)

	case 1:

		releases, err := s.GetServiceVersions(serviceName)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()
		fi, fVersion, err := ui.Select("Select first version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		copy(versions[fi:], versions[fi+1:]) // shift valuesafter the indexwith a factor of 1
		versions[len(versions)-1] = ""       // remove element
		versions = versions[:len(versions)-1]

		_, sVersion, err := ui.Select("Select last version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		changelog, err := s.GetChangelog(serviceName, fVersion, sVersion)
		if err != nil {
			return fmt.Errorf("getting changelog %w", err)
		}
		fmt.Print(changelog)

	case 2:
		return nil
	}

	return ServiceOptionsUI(ctx, s, serviceName)
}
