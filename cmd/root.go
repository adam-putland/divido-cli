/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"context"

	"github.com/adam-putland/divido-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/manifoldco/promptui"
)

var cfgFile string

var options = []string{
	"show services deployed in an environment",
	"show services in a helm chart",
	"diff between helm charts",
	"generate changelog between two given helm charts ",
	"bump a service in a helm chart",
	"bump a helm chart in an environment",
	"override/remove override a service in an environment",
	"undo / redo last commands",
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "divido-cli",
	Short: "A cli for Divido devs",
	Long: `This cli provides tools for deploying services, updating helm charts and updating environments`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := internal.NewGithubClient(ctx, viper.GetString("GITHUB_TOKEN"))

		prompt := promptui.Select{
			Label: "Select Option",
			Items: options,
		}

		index,_, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch index {
		case 0:
			fmt.Println(client.Client.Repositories.List(ctx, "dividohq", nil))


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
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".divido-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
