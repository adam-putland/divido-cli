/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"github.com/adam-putland/divido-cli/internal"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var config internal.Config

var options = []string{
	"show services deployed in an environment",
	"show services in a helm chart",
	"diff between helm charts",
	"generate changelog between two given helm charts ",
	"bump a service in a helm chart",
	"bump a helm chart in an environment",
	"override/remove a service in an environment",
	"undo / redo last commands",
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "divido-cli",
	Short: "A cli for Divido devs",
	Long:  `This cli provides tools for deploying services, updating helm charts and updating environments`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		deployer := internal.NewDeployer(ctx, &config, viper.GetString("GITHUB_TOKEN"))

		index, _, err := internal.Select("Select Option", options)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch index {
		case 0:
			index, _, err := internal.Select("Select platform", config.ListPlatform())
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			envIndex, _, err := internal.Select("Select env", config.ListEnvironments(index))
			if err != nil {
				return
			}

			services, err := deployer.GetEnvServices(ctx, index, envIndex)
			if err != nil {
				return
			}
			fmt.Printf("data: %s\n", services)
		case 1:
			index, _, err := internal.Select("Select platform", config.ListPlatform())
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}
			services, err := deployer.GetLatestChartServices(ctx, index)
			if err != nil {
				return
			}
			fmt.Printf("data: %s\n", services)
		}
	},
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

	config = internal.Config{}
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}
}
