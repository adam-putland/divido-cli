/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/adam-putland/divido-cli/internal"
	"github.com/adam-putland/divido-cli/internal/ui"
	"github.com/sarulabs/di"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var options = []string{
	"Services query",
	"Helm query",
	"Environments query",
	"Exit",
}

/// Service query -> prompts for input and shows the service, lists versions, click on version to get info, commit message, url to commit
// Helm query -> prompt for platform (divido) ->
// View -> list of versions -> select one version to see info (services, overrides, etc)
//Diff -> list of versions -> choose two and it generates diff + option to make changelogs (JIRA API)
//Bump -> show current version and give option to bump a service (could type in service and it shows versions you can choose)
// Environment query -> prompt for platform, env -> show hc, overrides, services
// undo / redo

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "divido-cli",
	Short: "A cli for Divido devs",
	Long:  `This cli provides tools for deploying services, updating helm charts and updating environments`,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := internal.CreateApp(ctx)
		if app == nil {
			return errors.New("error generation application")
		}
		return Run(ctx, *app)
	},
}

func Run(ctx context.Context, app di.Container) error {
	index, _, err := ui.Select("Select Option", options)
	if err != nil {
		return fmt.Errorf("select failed %v", err)
	}

	var errUI error
	switch index {
	case 0:
		errUI = ServiceUI(ctx, app)
	case 1:
		HelmUI(app)
	case 2:
		errUI = EnvUI(ctx, app)
	case 3:
		return nil
	}

	if errUI != nil {
		fmt.Println(errUI)
	}
	return Run(ctx, app)

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.divido-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".divido-cli" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)

		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
