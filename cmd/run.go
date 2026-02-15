package cmd

import (
	"strings"

	"github.com/amieldelatorre/shurl/internal"
	"github.com/spf13/cobra"
)

var configFilePath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the url shortener application",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		configFilePath = strings.TrimSpace(configFilePath)
		app := internal.NewApp(configFilePath)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&configFilePath, "filepath", "f", "", "The path to the config file, accepted file types are: .env, .ini, .toml, .yaml, .json")
}
