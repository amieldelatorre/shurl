package db

import (
	"context"

	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/google/uuid"
)

type DbContext interface {
	GetDatabaseVersion() int64
	CreateShortUrl(ctx context.Context, req types.CreateShortUrl, idempotencyKey uuid.UUID, request_hash string) (*types.ShortUrl, error)
	GetShortUrlById(ctx context.Context, id uuid.UUID) (*types.ShortUrl, error)
	GetShortUrlBySlug(ctx context.Context, slug string) (*types.ShortUrl, error)
	CreateUser(ctx context.Context, idempotencyKey uuid.UUID, requestHash string, req types.CreateUserRequest) (*types.User, error)
}
