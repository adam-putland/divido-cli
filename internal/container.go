package internal

import (
	"context"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/sarulabs/di"
	"github.com/spf13/viper"
)

func CreateApp(ctx context.Context) *di.Container {
	builder, _ := di.NewBuilder()

	err := builder.Add([]di.Def{
		{
			Name:  "github",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return github.NewGithubClient(ctx, viper.GetString("GITHUB_TOKEN")), nil
			},
			Close: nil},
		{
			Name:  "config",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				config := models.Config{}
				if err := viper.Unmarshal(&config); err != nil {
					return nil, err
				}
				return &config, nil
			},
			Close: nil},
		{
			Name:  "service",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return service.New(ctn.Get("github").(*github.GithubClient), ctn.Get("config").(*models.Config)), nil
			},
			Close: nil},
	}...)
	if err != nil {
		return nil
	}

	c := builder.Build()
	return &c
}
