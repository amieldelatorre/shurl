package cmd

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/amieldelatorre/shurl/internal"
	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the url shortener application",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configFilePath = strings.TrimSpace(configFilePath)
		tempLogger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)

		config, err := config.LoadConfig(configFilePath)
		if err != nil {
			tempLogger.ErrorExit(ctx, err.Error())
		}

		app := internal.NewApp(ctx, config)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
