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
	"os"
	"strings"
	"sync"
)

var helmOptions = []string{
	"Info",
	"Update Service(s) Versions",
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

		services := make([]*models.Service, 0, len(plat.Services))
		for _, ser := range plat.Services {
			services = append(services, ser)
		}

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
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}

	case 1:

		err = BumpServicesUI(ctx, s, platConfig)
		if err != nil {
			return err
		}

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

func BumpServicesUI(ctx context.Context, s *service.Service, platConfig *models.PlatformConfig) error {

	plat, err := s.GetPlat(ctx, platConfig.HelmChartRepo)
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
		fmt.Printf("Prompt failed %v", err)
		os.Exit(1)
	}

	selectedServices := make([]*models.ServiceUpdated, 0, len(selected))

	wg := sync.WaitGroup{}

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
			fmt.Printf("Prompt failed %v", err)
			os.Exit(1)
		}
	}

	config := s.GetConfig()
	githubDetails := github.WithBumpServices(&config.Github)
	return GithubUI(ctx, s, githubDetails, platConfig, selectedServices)

}

func GithubUI(ctx context.Context, s *service.Service, gd *github.Commit, platConfig *models.PlatformConfig, services []*models.ServiceUpdated) error {
	fmt.Printf("Github Details \n%s", gd)

	options := []string{
		"Change Author Name",
		"Change Author Email",
		"Change Commit Message",
		"Change Branch",
		"Continue",
		"Back",
	}

	if !platConfig.DirectCommit {
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

		err = s.UpdateServicesVersions(ctx, platConfig, gd, services)
		if err != nil {
			return fmt.Errorf("error updating services %w", err)
		}
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
	return GithubUI(ctx, s, gd, platConfig, services)
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
		//todo

	case 3:
		return nil
	}

	return VersionsUI(ctx, s, diff)
}
