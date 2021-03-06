package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/sarulabs/di"
	"strings"
)

var envOptions = util.Options{
	"Show Services and/or overrides",
	"Update Helm version",
	"Update Services via overrides (In development)",
}

func EnvUI(ctx context.Context, app di.Container) error {
	s := app.Get("service").(*service.Service)
	cfg := s.GetConfig()
	platIndex, _, err := util.Select("Select platform", cfg.ListPlatform())
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	envI, _, err := util.Select("Select env", cfg.ListEnvironments(platIndex))
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	env, err := s.GetEnv(ctx, platIndex, envI)
	if err != nil {
		return err
	}

	fmt.Println(env)

	return EnvOptionsUI(ctx, s, env, &cfg.Github, platIndex)
}

func EnvOptionsUI(ctx context.Context, s *service.Service, env *models.Environment, ghCfg *models.GithubConfig, platIndex int) error {

	option, _, err := util.Select(SelectOptionMsg, envOptions.WithBackOption())
	if err != nil {
		return fmt.Errorf(PromptFailedMsg, err)
	}

	switch option {
	case 0:
		err := s.LoadEnvServices(ctx, env, platIndex)
		if err != nil {
			return fmt.Errorf("loading environment services and overrides %w", err)
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

		if len(env.Overrides) > 0 {

			overrides := env.Overrides.ToArray()
			searcherO := func(input string, index int) bool {
				s := overrides[index]
				name := strings.Replace(strings.ToLower(s.Name), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)
				return strings.Contains(name, input)
			}

			prompt := util.MultiSelect{
				Label:     "Overrides:",
				Items:     overrides,
				Templates: templates,
				Size:      8,
				Searcher:  searcherO,
				HideHelp:  false,
			}
			_, err = prompt.Run()
			if err != nil {
				return fmt.Errorf(PromptFailedMsg, err)
			}
		}

		if len(env.Services) > 0 {

			services := env.Services.ToArray()
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
		}

	case 1:

		if env.OnlyOverrides {
			fmt.Printf("%s only updated via overrides (select show services)", env.Name)
			break
		}

		fmt.Printf("Current Version: %s", env.HelmChartVersion)

		releases, err := s.GetHelmVersions(ctx, env, platIndex)
		if err != nil {
			return fmt.Errorf("getting service versions %w", err)
		}

		versions := releases.Versions()
		_, fVersion, err := util.Select("Select version", util.Options(versions))
		if err != nil {
			return fmt.Errorf("Prompt failed %v\n", err)
		}

		githubDetails := github.WithBumpHC(ghCfg, fVersion)
		err = BumpHelmUI(ctx, s, env, githubDetails, fVersion)
		if err != nil {
			fmt.Println(err)
		}

	case 2:

		//
		fmt.Println("In development")
	case 3:
		return nil
	}

	return EnvOptionsUI(ctx, s, env, ghCfg, platIndex)
}

func BumpHelmUI(ctx context.Context, s *service.Service, env *models.Environment, gd *github.Commit, version string) error {
	fmt.Printf("Github Details \n%s", gd)

	options := util.Options{
		"Change Author Name",
		"Change Author Email",
		"Change Commit Message",
		"Change Branch",
		"Continue",
	}.WithBackOption()

	if !env.DirectCommit {
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
			return fmt.Errorf(PromptFailedMsg, err)
		}

	case 7:
		gd.PullRequestDescription, err = util.PromptWithDefault("Enter Pull request description", gd.PullRequestDescription)
		if err != nil {
			return fmt.Errorf(PromptFailedMsg, err)
		}

	}
	return BumpHelmUI(ctx, s, env, gd, version)
}
