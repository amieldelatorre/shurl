package valkey_cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

type ValkeyCacheContext struct {
	logger    utils.CustomJsonLogger
	dbContext db.DbContext
	client    *glide.Client
}

const (
	DB_VERSION_KEY        = "goose_db_version_max_id"
	SHORT_URL_ID_PREFIX   = "short_url:id:"
	SHORT_URL_SLUG_PREFIX = "short_url:slug:"
	USER_EMAIL_PREFIX     = "shurl_user:email:"
)

func NewValkeyCacheContext(logger utils.CustomJsonLogger, client *glide.Client, dbContext db.DbContext) *ValkeyCacheContext {
	return &ValkeyCacheContext{logger: logger, client: client, dbContext: dbContext}
}

func (v *ValkeyCacheContext) Ping(ctx context.Context) error {
	_, err := v.client.Ping(ctx)
	return err
}

func (v *ValkeyCacheContext) Close() {
	v.client.Close()
}

func (v *ValkeyCacheContext) GetDatabaseVersion(ctx context.Context) (int64, error) {
	resStr, err := v.getKey(ctx, DB_VERSION_KEY)
	if err != nil {
		v.logger.Error(ctx, "error getting database version from cache")
	}

	if resStr != nil {
		return strconv.ParseInt(*resStr, 10, 64)
	}

	version, err := v.dbContext.GetDatabaseVersion(ctx)
	if err != nil {
		return version, err
	}

	versionStr := strconv.FormatInt(version, 10)
	err = v.setKey(ctx, DB_VERSION_KEY, versionStr)
	if err != nil {
		v.logger.Error(ctx, "could not set database version in valkey")
	}

	return version, nil

}

func (v *ValkeyCacheContext) CreateShortUrl(ctx context.Context, req types.CreateShortUrl, idempotencyKey uuid.UUID, request_hash string) (*types.ShortUrl, error) {
	return v.dbContext.CreateShortUrl(ctx, req, idempotencyKey, request_hash)
}

func (v *ValkeyCacheContext) GetShortUrlById(ctx context.Context, id uuid.UUID) (*types.ShortUrl, error) {
	cacheKey := SHORT_URL_ID_PREFIX + id.String()
	resStr, err := v.getKey(ctx, cacheKey)
	if err != nil {
		v.logger.Error(ctx, "error getting short url by id from cache")
	}

	if resStr != nil {
		var r types.ShortUrl
		err = json.Unmarshal([]byte(*resStr), &r)
		if err != nil {
			return nil, err
		}

		return &r, nil
	}

	shortUrl, err := v.dbContext.GetShortUrlById(ctx, id)
	if err != nil {
		return nil, err
	}

	if shortUrl == nil {
		return nil, err
	}

	strData, err := json.Marshal(shortUrl)
	if err != nil {
		v.logger.Error(ctx, "Could not marshal data for valkey")
		return shortUrl, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set short url in valkey")
	}
	return shortUrl, nil
}

func (v *ValkeyCacheContext) GetShortUrlBySlug(ctx context.Context, slug string) (*types.ShortUrl, error) {
	cacheKey := SHORT_URL_SLUG_PREFIX + slug
	resStr, err := v.getKey(ctx, cacheKey)
	if err != nil {
		v.logger.Error(ctx, "error getting short url by slug from cache")
	}

	if resStr != nil {
		var r types.ShortUrl
		err = json.Unmarshal([]byte(*resStr), &r)
		if err != nil {
			return nil, err
		}

		return &r, nil
	}

	shortUrl, err := v.dbContext.GetShortUrlBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if shortUrl == nil {
		return nil, err
	}

	strData, err := json.Marshal(shortUrl)
	if err != nil {
		v.logger.Error(ctx, "could not marshal short url for valkey")
		return shortUrl, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set short url in valkey")
	}

	return shortUrl, nil
}

func (v *ValkeyCacheContext) CreateUser(ctx context.Context, idempotencyKey uuid.UUID, requestHash string, req types.CreateUserRequest) (*types.User, error) {
	return v.dbContext.CreateUser(ctx, idempotencyKey, requestHash, req)
}
func (v *ValkeyCacheContext) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	cacheKey := USER_EMAIL_PREFIX + email
	resStr, err := v.getKey(ctx, cacheKey)
	if err != nil {
		v.logger.Error(ctx, "error getting user by email from cache")
	}

	if resStr != nil {
		var r types.User
		err = json.Unmarshal([]byte(*resStr), &r)
		if err != nil {
			return nil, err
		}

		return &r, nil
	}

	user, err := v.dbContext.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, err
	}

	strData, err := json.Marshal(user)
	if err != nil {
		v.logger.Error(ctx, "could not marshal user for valkey")
		return user, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set user in valkey")
	}

	return user, nil
}

func (v *ValkeyCacheContext) DeleteExpiredIdempotencyKeys(ctx context.Context) (int, error) {
	return v.dbContext.DeleteExpiredIdempotencyKeys(ctx)
}

func (v *ValkeyCacheContext) DeleteExpiredIdempotencyKeysBatched(ctx context.Context, batchSize int) (int, error) {
	return v.dbContext.DeleteExpiredIdempotencyKeysBatched(ctx, batchSize)
}

func (v *ValkeyCacheContext) DeleteExpiredShortUrls(ctx context.Context) (int, error) {
	return v.dbContext.DeleteExpiredShortUrls(ctx)
}

func (v *ValkeyCacheContext) DeleteExpiredShortUrlsBatched(ctx context.Context, batchSize int) (int, error) {
	return v.dbContext.DeleteExpiredShortUrlsBatched(ctx, batchSize)
}

func (v *ValkeyCacheContext) getKey(ctx context.Context, key string) (*string, error) {
	value, err := v.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if value.IsNil() {
		return nil, nil
	}

	result := value.Value()
	return &result, nil
}

func (v *ValkeyCacheContext) setKey(ctx context.Context, key string, value string) error {
	_, err := v.client.SetWithOptions(ctx, key, value, options.SetOptions{
		Expiry: options.NewExpiryIn(time.Duration(300) * time.Second),
	})
	return err
}
