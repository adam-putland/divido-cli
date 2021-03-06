package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/sarulabs/di"
	"strings"
)

var serviceOptions = util.Options{
	"Versions",
	"Generate Changelog",
}

func ServiceUI(ctx context.Context, app di.Container) error {
	s := app.Get("service").(*service.Service)
	serviceName, err := util.Prompt("Enter service")
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	serv, err := s.GetLatest(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("getting service %w", err)
	}
	fmt.Println(serv)

	return ServiceOptionsUI(ctx, s, serviceName)
}

func ServiceOptionsUI(ctx context.Context, s *service.Service, serviceName string) error {
	option, _, err := util.Select(SelectOptionMsg, serviceOptions.WithBackOption())
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	switch option {
	case 0:
		releases, err := s.GetRepoReleases(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()
		_, _, err = util.SelectWithSearch("Versions", util.Options(releases.Versions()), func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

	case 1:

		releases, err := s.GetRepoReleases(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()

		fi, fVersion, err := util.SelectWithSearch("Select first version", util.Options(versions), func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

		versions.Remove(fi)

		_, sVersion, err := util.SelectWithSearch("Select last version", util.Options(versions), func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
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
