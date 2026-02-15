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

func GetDatabaseContext(ctx context.Context, config config.Config, logger utils.CustomJsonLogger) DbContext {
	dbMigrations, err := postgresql.NewPostgresDatabaseMigrations(ctx, config, logger)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	if err = goose.SetDialect(dbMigrations.GetGooseDialect()); err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	currentDbVersion, err := goose.GetDBVersion(dbMigrations.GetDb())
	logger.Info(ctx, "Current database version is: %v", "currentDbVersion", currentDbVersion)

	goose.SetBaseFS(dbMigrations.GetEmbedMigrations())
	if err = goose.Up(dbMigrations.GetDb(), "migrations"); err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	logger.Info(ctx, "Successfully connected to the database")

	dbContext := postgresql.NewPostreSQLContext(dbMigrations.DbPool)
	return dbContext
}
