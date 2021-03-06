package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/manifoldco/promptui"
	"github.com/sarulabs/di"
	"strings"
	"sync"
)

var helmOptions = util.Options{
	"Info",
	"Update Service(s) Versions",
	"Compare Versions",
}

func HelmUI(ctx context.Context, app di.Container) error {

	s := app.Get("service").(*service.Service)
	cfg := s.GetConfig()
	platIndex, _, err := util.Select("Select platform", cfg.ListPlatform())
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	platConfig := cfg.GetPlatform(platIndex)

	return HelmOptionsUI(ctx, s, platConfig)

}

func HelmOptionsUI(ctx context.Context, s *service.Service, platCfg *models.PlatformConfig) error {

	latest, err := s.GetLatest(ctx, platCfg.HelmChartRepo)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Latest Release: \n%s", latest)

	option, _, err := util.Select(SelectOptionMsg, helmOptions.WithBackOption())
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	switch option {
	case 0:

		plat, err := s.GetPlat(ctx, platCfg.HelmChartRepo)
		if err != nil {
			return fmt.Errorf("getting services in hlm %w", err)
		}
		fmt.Println(plat)

		services := plat.Services.ToArray()

		templates := &util.MultiSelectTemplates{
			Label:      "{{ . }}",
			Selected:   "\U00002388 {{ .Name | cyan }}: {{ .Version | cyan }}",
			Unselected: "  {{ .Name | cyan }}: {{ .Version | cyan }}",
			Help: fmt.Sprintf(`{{ "Use the arrow keys to navigate:" | faint }} {{ .NextKey | faint }} ` +
				`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
				`{{ if .Search }}{{ " (" | faint }}{{ .SearchKey | faint }} {{ "to search)" | faint }} {{ end }}` +
				`{{ " (Press enter to quit)" | faint }}`),
		}

		searcher := func(input string, index int) bool {
			s := services[index]
			name := strings.Replace(strings.ToLower(s.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		}

		prompt := util.MultiSelect{
			Label:     "Services:",
			Items:     services,
			Templates: templates,
			Size:      8,
			Searcher:  searcher,
			HideHelp:  false,
		}

		_, err = prompt.Run()
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

	case 1:

		err = BumpServicesUI(ctx, s, platCfg)
		if err != nil {
			return err
		}

	case 2:
		releases, err := s.GetRepoReleases(ctx, platCfg.HelmChartRepo)
		if err != nil {
			return fmt.Errorf("getting platform versions %w", err)
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
		_, lVersion, err := util.SelectWithSearch("Select last version", util.Options(versions), func(input string, index int) bool {
			s := versions[index]
			name := strings.Replace(strings.ToLower(s), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		})
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

		diff, err := s.ComparePlatReleasesByVersion(ctx, platCfg, releases, fVersion, lVersion)
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

	return HelmOptionsUI(ctx, s, platCfg)
}

func BumpServicesUI(ctx context.Context, s *service.Service, platCfg *models.PlatformConfig) error {

	plat, err := s.GetPlat(ctx, platCfg.HelmChartRepo)
	if err != nil {
		return fmt.Errorf("getting services in hlm %w", err)
	}

	services := make([]*models.ServiceUpdated, 0, len(plat.Services))
	for _, ser := range plat.Services {
		services = append(services, &models.ServiceUpdated{Service: ser})
	}

	templates := &util.MultiSelectTemplates{
		Label:      "{{ . }}",
		Selected:   "\U00002388 {{ .Service.Name | cyan }}: {{ if .NewVersion }}{{ .Service.Version | red }} -> {{ .NewVersion | green }}{{ else }}{{ .Service.Version | cyan }}{{ end }}",
		Unselected: "  {{ .Service.Name | cyan }}: {{ if .NewVersion }}{{ .Service.Version | red }} -> {{ .NewVersion | green }}{{ else }}{{ .Service.Version | cyan }}{{ end }}",
		Help: fmt.Sprintf(`{{ "Use the arrow keys to navigate:" | faint }} {{ .NextKey | faint }} ` +
			`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
			`{{ if .Search }}{{ " (" | faint }}{{ .SearchKey | faint }} {{ "to search)" | faint }} {{ end }}` +
			`{{ " (" | faint }}{{ .ToggleKey | faint }} {{ "to select)" | faint }}` +
			`{{ " (Press enter to quit and save)" | faint }}`),
	}

	searcher := func(input string, index int) bool {
		s := services[index]
		name := strings.Replace(strings.ToLower(s.Service.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := util.MultiSelect{
		Label:     "Select Services to update",
		Items:     services,
		Templates: templates,
		Size:      8,
		Searcher:  searcher,
		HideHelp:  false,
	}

	selected, err := prompt.Run()
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	selectedServices := make([]*models.ServiceUpdated, 0, len(selected))

	var wg sync.WaitGroup

	rowRes := make(map[int][]string, len(selected))

	for index, option := range selected {
		selectedServices = append(selectedServices, services[option])
		wg.Add(1)
		go func(index int, service *models.ServiceUpdated, w *sync.WaitGroup) {

			releases, _ := s.GetAvailableServiceReleases(ctx, service.Service)
			rowRes[index] = releases.Versions()
			w.Done()
		}(index, services[option], &wg)
	}
	fmt.Println("....Obtaining available versions....")
	wg.Wait()

	for index := range selectedServices {
		if versions, ok := rowRes[index]; ok && len(versions) > 0 {
			_, selectedServices[index].NewVersion, err = util.SelectWithAdd(fmt.Sprintf("%s current (%s)", selectedServices[index].Service.Name, selectedServices[index].Service.Version), versions)
		} else {
			selectedServices[index].NewVersion, err = util.PromptWithDefault(selectedServices[index].Service.Name, selectedServices[index].Service.Version)
		}
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}
	}

	cfg := s.GetConfig()
	githubDetails := github.WithBumpServices(&cfg.Github)
	return GithubUI(ctx, s, githubDetails, platCfg, selectedServices)

}

func GithubUI(ctx context.Context, s *service.Service, gd *github.Commit, platCfg *models.PlatformConfig, services []*models.ServiceUpdated) error {
	fmt.Printf("Github Details \n%s", gd)

	options := util.Options{
		"Change Author Name",
		"Change Author Email",
		"Change Commit Message",
		"Change Branch",
		"Continue",
	}.WithBackOption()

	if !platCfg.DirectCommit {
		fmt.Print(gd.PullRequestInfo())
		options = append(options, []string{"Change pull request title", "Change pull request description"}...)
	}

	githubC, _, err := util.Select(SelectOptionMsg, options)
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	switch githubC {
	case 0:
		gd.AuthorName, err = util.PromptWithDefault("Enter Author Name", gd.AuthorName)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

	case 1:
		gd.AuthorName, err = util.PromptWithDefault("Enter Author Email", gd.AuthorEmail)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}
	case 2:
		gd.AuthorName, err = util.PromptWithDefault("Enter Commit Message", gd.Message)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}
	case 3:
		gd.AuthorName, err = util.PromptWithDefault("Enter Branch", gd.Branch)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}
	case 4:

		err = s.UpdateServicesVersions(ctx, platCfg, gd, services)
		if err != nil {
			return fmt.Errorf("error updating services %w", err)
		}
		return nil
	case 5:
		return nil
	case 6:
		gd.PullRequestTitle, err = util.PromptWithDefault("Enter Pull request title", gd.PullRequestTitle)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

	case 7:
		gd.PullRequestDescription, err = util.PromptWithDefault("Enter Pull request description", gd.PullRequestDescription)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

	}
	return GithubUI(ctx, s, gd, platCfg, services)
}

func VersionsUI(ctx context.Context, s *service.Service, diff *models.Comparer) error {

	options := util.Options{
		"Show Changelogs",
		"Export Release",
		"Create Release Ticket (In development)",
	}

	option, _, err := util.Select(SelectOptionMsg, options.WithBackOption())
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	switch option {
	case 0:
		changelogs, err := s.GetChangelogsFromDiff(ctx, diff)
		if err != nil {
			return err
		}

		for key, changelog := range changelogs {
			fmt.Printf("\nService Repo: %s\n", key)
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
		//todo
		fmt.Println("In development")
	case 3:
		return nil
	}

	return VersionsUI(ctx, s, diff)
}
