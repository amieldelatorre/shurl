package cmd

import (
	"strings"

	"github.com/amieldelatorre/shurl/internal"
	"github.com/spf13/cobra"
)

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
}
