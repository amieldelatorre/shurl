package db

import (
	"context"

	"github.com/amieldelatorre/shurl/internal/types"
)

type DbContext interface {
	GetDatabaseVersion() int64
	CreateShortUrl(ctx context.Context, req types.CreateShortUrl) (*types.ShortUrl, error)
	GetShortUrlBySlug(ctx context.Context, slug string) (*types.ShortUrl, error)
}
