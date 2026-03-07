package db

import (
	"context"
	"database/sql"
	"embed"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db/postgresql"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/pressly/goose/v3"
)

type DatabaseMigrations interface {
	GetDb() *sql.DB
	GetEmbedMigrations() embed.FS
	GetGooseDialect() string
}

func GetDatabaseContext(ctx context.Context, config config.Config, logger utils.CustomJsonLogger, forceRunMigrations bool) DbContext {
	dbMigrations, err := postgresql.NewPostgresDatabaseMigrations(ctx, config, logger)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	migrationsFs, err := dbMigrations.GetEmbedMigrations()
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	provider, err := goose.NewProvider(
		dbMigrations.GetGooseDialect(),
		dbMigrations.GetDb(),
		migrationsFs,
	)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	currentDbVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	logger.Info(ctx, "current database version is: %v", "currentDbVersion", currentDbVersion)
	if *config.Database.RunMigrations || forceRunMigrations {
		logger.Info(ctx, "running migrations")
		applied, err := provider.Up(ctx)
		if err != nil {
			logger.ErrorExit(ctx, err.Error())
		}

		for _, a := range applied {
			logger.Info(ctx, "ran migration", "migration", a.String())
		}
	} else {
		logger.Info(ctx, "skipped migrations")
	}

	logger.Info(ctx, "Successfully connected to the database")

	dbContext := postgresql.NewPostreSQLContext(logger, dbMigrations.DbPool)
	return dbContext
}
