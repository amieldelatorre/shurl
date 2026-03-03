/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/spf13/cobra"
)

// runMigrationsCmd represents the runMigrations command
var runMigrationsCmd = &cobra.Command{
	Use:   "run-migrations",
	Short: "Run migrations",
	Long:  `Run migrations`,
	Run: func(cmd *cobra.Command, args []string) {
		tempLogger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15)*time.Second)
		defer cancel()
		configFilePath = strings.TrimSpace(configFilePath)
		conf, err := config.LoadConfig(configFilePath)
		if err != nil {
			fmt.Println("Couldn't run migrations error:", err.Error())
		}

		dbContext := db.GetDatabaseContext(ctx, *conf, tempLogger, true)
		dbContext.Close()
	},
}

func init() {
	rootCmd.AddCommand(runMigrationsCmd)
}
