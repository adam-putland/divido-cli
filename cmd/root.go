/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/adam-putland/divido-cli/internal"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/adam-putland/divido-cli/internal/service"
	"github.com/adam-putland/divido-cli/internal/ui"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var config models.Config

var options = []string{
	"Services query",
	"Helm query",
	"Environments query",
	"undo / redo last commands",
}

var serviceOptions = []string{
	"Versions",
	"Generate Changelog",
	"Back",
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "divido-cli",
	Short: "A cli for Divido devs",
	Long:  `This cli provides tools for deploying services, updating helm charts and updating environments`,

	Run: func(cmd *cobra.Command, args []string) {
		app := internal.CreateApp()

		index, _, err := ui.Select("Select Option", options)
		if err != nil {
			fmt.Printf("Select failed %v\n", err)
			return
		}

		switch index {
		case 0:
			s := app.Get("service").(*service.Service)
			in, err := ui.Prompt("Enter service:")
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			serv, err := s.GetServiceLatest(in)
			if err != nil {
				fmt.Printf("Error getting service %v\n", err)
				return
			}

			fmt.Print("\033[H\033[2J")
			fmt.Printf("%s\nlatest version:%s\nURL:%s\n", serv.Name, serv.Version, serv.URL)
			i ,_, err := ui.Select("Choose option", serviceOptions)

			switch i {
			case 0:
				versions, err := s.GetServiceVersions(in)
				if err != nil {
					fmt.Printf("Error getting service versions %v\n", err)
					return
				}


				fmt.Print(versions)


			case 1:
				//fmt.Println(s.GetChangelog(in,))
			}
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
}
