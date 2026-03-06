package db

import (
	"context"

	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/google/uuid"
)

type DbContext interface {
	Ping(ctx context.Context) error
	GetDatabaseVersion(ctx context.Context) (int64, error)
	CreateShortUrl(ctx context.Context, req types.CreateShortUrl, idempotencyKey uuid.UUID, request_hash string) (*types.ShortUrl, error)
	GetShortUrlsByUserId(ctx context.Context, userId uuid.UUID, page int, size int) ([]types.ShortUrl, error)
	GetShortUrlById(ctx context.Context, id uuid.UUID, excludeExpired bool) (*types.ShortUrl, error)
	GetShortUrlBySlug(ctx context.Context, slug string, excludeExpired bool) (*types.ShortUrl, error)
	DeleteShortUrlById(ctx context.Context, userId uuid.UUID, shortUrlId uuid.UUID) (types.DeleteShortUrlResult, error)
	CreateUser(ctx context.Context, idempotencyKey uuid.UUID, requestHash string, req types.CreateUserRequest) (*types.User, error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	DeleteExpiredIdempotencyKeys(ctx context.Context) (int, error)
	DeleteExpiredIdempotencyKeysBatched(ctx context.Context, batchSize int) (int, error)
	DeleteExpiredShortUrls(ctx context.Context) (int, error)
	DeleteExpiredShortUrlsBatched(ctx context.Context, batchSize int) (int, error)
	Close()
}
