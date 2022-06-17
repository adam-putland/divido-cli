package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/ui"
	"github.com/sarulabs/di"
)

var envOptions = []string{
	"Info",
	"Bump helm",
	"Back",
}

func EnvUI(ctx context.Context, app di.Container) {
	//s := app.Get("service").(*service.Service)
	config := app.Get("config").(*models.Config)
	index, _, err := ui.Select("Select platform", config.ListPlatform())
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	_, _, err = ui.Select("Select env", config.ListEnvironments(index))
	if err != nil {
		return
	}

	//services, err := s.GetEnvServices(ctx, index, envIndex)
	//if err != nil {
	//	return
	//}
	//fmt.Printf("data: %s\n", services)

}
