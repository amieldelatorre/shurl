package postgresql

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations
var embedMigrations embed.FS

type PostgresDatabaseMigrations struct {
	DbPool *pgxpool.Pool
}

func NewPostgresDatabaseMigrations(ctx context.Context, config config.Config, logger utils.CustomJsonLogger) (*PostgresDatabaseMigrations, error) {
	connstring := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?target_session_attrs=read-write&connect_timeout=5", config.Database.Username, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name)

	dbPool, err := pgxpool.New(context.Background(), connstring)
	if err != nil {
		return nil, err
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgresDatabaseMigrations{DbPool: dbPool}, nil
}

func (m *PostgresDatabaseMigrations) GetDb() *sql.DB {
	stdDbPool := stdlib.OpenDBFromPool(m.DbPool)
	return stdDbPool
}

func (m *PostgresDatabaseMigrations) GetEmbedMigrations() embed.FS {
	return embedMigrations
}

func (m *PostgresDatabaseMigrations) Close() {
	m.DbPool.Close()
}

func (m *PostgresDatabaseMigrations) GetGooseDialect() string {
	return string(goose.DialectPostgres)
}
