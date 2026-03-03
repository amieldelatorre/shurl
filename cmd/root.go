package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var configFilePath string

var rootCmd = &cobra.Command{
	Use:   "shurl",
	Short: "For url shortener",
	Long:  `A url shortener application.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "filepath", "f", "", "The path to the config file, accepted file types are: .env, .ini, .toml, .yaml, .json")
}
