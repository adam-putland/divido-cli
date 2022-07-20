package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/sarulabs/di"
	"os"
	"strings"
)

var serviceOptions = []string{
	"Versions",
	"Generate Changelog",
	"Back",
}

func ServiceUI(ctx context.Context, app di.Container) error {
	s := app.Get("service").(*service.Service)
	serviceName, err := util.Prompt("Enter service")
	if err != nil {
		fmt.Printf(PromptFailedMsg, err)
		os.Exit(1)
	}

	serv, err := s.GetLatest(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("getting service %w", err)
	}
	fmt.Println(serv)

	return ServiceOptionsUI(ctx, s, serviceName)
}

func ServiceOptionsUI(ctx context.Context, s *service.Service, serviceName string) error {
	option, _, err := util.Select(SelectOptionMsg, serviceOptions)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	switch option {
	case 0:
		releases, err := s.GetRepoReleases(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()
		_, _, err = util.SelectWithSearch("Versions", releases.Versions(), func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

	case 1:

		releases, err := s.GetRepoReleases(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()

		fi, fVersion, err := util.SelectWithSearch("Select first version", versions, func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		versions.Remove(fi)

		_, sVersion, err := util.SelectWithSearch("Select last version", versions, func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}

		changelog, err := s.GetChangelog(ctx, serviceName, releases.GetReleaseByVersion(fVersion), releases.GetReleaseByVersion(sVersion))
		if err != nil {
			return fmt.Errorf("getting changelog %w", err)
		}
		fmt.Print(changelog)

	case 2:
		return nil
	}

	return ServiceOptionsUI(ctx, s, serviceName)
}
