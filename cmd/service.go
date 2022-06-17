package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
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

func ServiceUI(ctx context.Context, app di.Container) {
	s := app.Get("service").(*service.Service)
	in, err := ui.Prompt("Enter service")
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	serv, err := s.GetServiceLatest(in)
	if err != nil {
		fmt.Printf("Error getting service %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\nlatest version: %s\nURL: %s\n", serv.Name, serv.Version, serv.URL)

	ServiceOptionsUI(ctx, app, s, in)
}

func ServiceOptionsUI(ctx context.Context, app di.Container, s *service.Service, in string) {
	i, _, err := ui.Select("Choose option", serviceOptions)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	switch i {
	case 0:
		versions, err := s.GetServiceVersions(in)
		if err != nil {
			fmt.Printf("Error getting service versions %v\n", err)
			os.Exit(1)
		}
		fmt.Print(versions)
		ctx = context.WithValue(ctx, fmt.Sprintf("%s-releases", in), versions)

	case 1:
		releases, ok := ctx.Value(fmt.Sprintf("%s-releases", in)).(models.Releases)
		if !ok {
			releases, err = s.GetServiceVersions(in)
			if err != nil {
				fmt.Printf("Error getting service versions %v\n", err)
				os.Exit(1)
			}
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

		_, sVersion, err := ui.Select("Select first version", versions)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		changelog, err := s.GetChangelog(in, fVersion, sVersion)
		if err != nil {
			fmt.Printf("Error getting service versions %v\n", err)
			os.Exit(1)
		}
		fmt.Print(changelog)

	case 2:
		Run(ctx, app)
		return
	}
	ServiceOptionsUI(ctx, app, s, in)
}
