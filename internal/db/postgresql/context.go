package postgresql

import (
	"context"
	"errors"

	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQLContext struct {
	dbPool *pgxpool.Pool
}

func NewPostreSQLContext(dbPool *pgxpool.Pool) *PostgreSQLContext {
	return &PostgreSQLContext{dbPool: dbPool}
}

func (p *PostgreSQLContext) GetDatabaseVersion() int64 {
	return 2
}

func (p *PostgreSQLContext) CreateShortUrl(ctx context.Context, req types.CreateShortUrl) (*types.ShortUrl, error) {
	tx, err := p.dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var newShortUrl types.ShortUrl
	err = p.dbPool.QueryRow(ctx,
		`INSERT INTO short_urls (id, destination_url, slug, created_at)
		 VALUES ($1, $2, $3, NOW())
		 RETURNING id, destination_url, slug, created_at`,
		req.Id, req.DestinationUrl, req.Slug).Scan(
		&newShortUrl.Id, &newShortUrl.DestinationUrl, &newShortUrl.Slug, &newShortUrl.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	return &newShortUrl, err
}

func (p *PostgreSQLContext) GetShortUrlBySlug(ctx context.Context, slug string) (*types.ShortUrl, error) {

	var shortUrl types.ShortUrl

	// slug should be unique
	err := p.dbPool.QueryRow(ctx, `SELECT * FROM short_urls WHERE slug = $1`, slug).Scan(
		&shortUrl.Id, &shortUrl.DestinationUrl, &shortUrl.Slug, &shortUrl.CreatedAt,
	)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &shortUrl, err
}
