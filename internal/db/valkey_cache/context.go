package valkey_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

type ValkeyCacheContext struct {
	logger    utils.CustomJsonLogger
	dbContext db.DbContext
	client    *glide.Client
}

const (
	DB_VERSION_KEY               = "goose_db_version_max_id"
	USER_EMAIL_PREFIX            = "shurl_user:email::"
	CACHE_DOUBLE_DELETE_SLEEP_MS = 250
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
		v.logger.Error(ctx, "error getting database version from cache", "error", err.Error())
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
		v.logger.Error(ctx, "could not set database version in valkey", "error", err.Error())
	}

	return version, nil

}

func (v *ValkeyCacheContext) CreateShortUrl(ctx context.Context, req types.CreateShortUrl, idempotencyKey uuid.UUID, request_hash string) (*types.ShortUrl, error) {
	delKeys := func() {
		if req.UserId != nil {
			err := v.delUserShortUrlQueries(ctx, getShortUrlsByUserIdCachePrefix(*req.UserId)+"*")
			if err != nil {
				v.logger.Error(ctx, "couldn't unlink keys from valkey", "error", err.Error())
			}
		}
	}

	delKeys()
	res, resErr := v.dbContext.CreateShortUrl(ctx, req, idempotencyKey, request_hash)
	time.Sleep(CACHE_DOUBLE_DELETE_SLEEP_MS * time.Millisecond)
	delKeys()

	return res, resErr
}

func (v *ValkeyCacheContext) GetShortUrlById(ctx context.Context, id uuid.UUID, excludeExpired bool) (*types.ShortUrl, error) {
	cacheKey := getShortUrlByIdCachePrefix(id)
	resStr, err := v.getKey(ctx, cacheKey)
	if err != nil {
		v.logger.Error(ctx, "error getting short url by id from cache", "error", err.Error())
	}

	if resStr != nil {
		var r types.ShortUrl
		err = json.Unmarshal([]byte(*resStr), &r)
		if err != nil {
			return nil, err
		}

		return &r, nil
	}

	shortUrl, err := v.dbContext.GetShortUrlById(ctx, id, excludeExpired)
	if err != nil {
		return nil, err
	}

	if shortUrl == nil {
		return nil, err
	}

	strData, err := json.Marshal(shortUrl)
	if err != nil {
		v.logger.Error(ctx, "Could not marshal data for valkey", "error", err.Error())
		return shortUrl, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set short url in valkey", "error", err.Error())
	}
	return shortUrl, nil
}

func (v *ValkeyCacheContext) GetShortUrlBySlug(ctx context.Context, slug string, excludeExpired bool) (*types.ShortUrl, error) {
	cacheKey := getShortUrlBySlugCachePrefix(slug)
	resStr, err := v.getKey(ctx, cacheKey)
	if err != nil {
		v.logger.Error(ctx, "error getting short url by slug from cache", "error", err.Error())
	}

	if resStr != nil {
		var r types.ShortUrl
		err = json.Unmarshal([]byte(*resStr), &r)
		if err != nil {
			return nil, err
		}

		return &r, nil
	}

	shortUrl, err := v.dbContext.GetShortUrlBySlug(ctx, slug, excludeExpired)
	if err != nil {
		return nil, err
	}

	if shortUrl == nil {
		return nil, err
	}

	strData, err := json.Marshal(shortUrl)
	if err != nil {
		v.logger.Error(ctx, "could not marshal short url for valkey", "error", err.Error())
		return shortUrl, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set short url in valkey", "error", err.Error())
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
		v.logger.Error(ctx, "error getting user by email from cache", "error", err.Error())
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
		v.logger.Error(ctx, "could not marshal user for valkey", "error", err.Error())
		return user, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set user in valkey", "error", err.Error())
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

func (v *ValkeyCacheContext) GetShortUrlsByUserId(ctx context.Context, userId uuid.UUID, size int, offset int) (types.GetShortUrlsResult, error) {
	cacheKey := getShortUrlsByUserIdCacheKey(userId, size, offset)

	resStr, err := v.getKey(ctx, cacheKey)
	if err != nil {
		v.logger.Error(ctx, "error getting short urls by user id from cache", "error", err.Error())
	}

	var userShortUrls types.GetShortUrlsResult
	if resStr != nil {
		err = json.Unmarshal([]byte(*resStr), &userShortUrls)
		if err != nil {
			return userShortUrls, err
		}

		return userShortUrls, nil
	}

	userShortUrls, err = v.dbContext.GetShortUrlsByUserId(ctx, userId, size, offset)
	if err != nil {
		return userShortUrls, err
	}

	strData, err := json.Marshal(userShortUrls)
	if err != nil {
		v.logger.Error(ctx, "could not short urls for valkey", "error", err.Error())
		return userShortUrls, nil
	}

	err = v.setKey(ctx, cacheKey, string(strData))
	if err != nil {
		v.logger.Error(ctx, "could not set short urls for user id in valkey", "error", err.Error())
	}

	return userShortUrls, nil
}

func (v *ValkeyCacheContext) DeleteShortUrlById(ctx context.Context, userId uuid.UUID, shortUrlId uuid.UUID) (types.DeleteShortUrlResult, error) {
	delKeys := func() {
		err := v.delKeys(ctx, []string{getShortUrlByIdCachePrefix(shortUrlId)})
		if err != nil {
			v.logger.Error(ctx, "couldn't delete keys from valkey", "error", err.Error())
		}
		err = v.delUserShortUrlQueries(ctx, getShortUrlsByUserIdCachePrefix(userId)+"*")
		if err != nil {
			v.logger.Error(ctx, "couldn't unlink keys from valkey", "error", err.Error())
		}
	}

	delKeys()
	result, resultErr := v.dbContext.DeleteShortUrlById(ctx, userId, shortUrlId)
	time.Sleep(CACHE_DOUBLE_DELETE_SLEEP_MS * time.Millisecond)
	delKeys()

	return result, resultErr
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

func (v *ValkeyCacheContext) delKeys(ctx context.Context, keys []string) error {
	deleted, err := v.client.Unlink(ctx, keys)
	if err != nil {
		return err
	}
	v.logger.Debug(ctx, "deleted keys from valkey", "count", deleted)
	return nil
}

func (v *ValkeyCacheContext) delUserShortUrlQueries(ctx context.Context, keyPrefix string) error {
	cursor := models.NewCursor()
	scanOpts := options.NewScanOptions().SetMatch(keyPrefix).SetCount(100)
	for {
		res, err := v.client.ScanWithOptions(ctx, cursor, *scanOpts)
		if err != nil {
			return err
		}

		keys := res.Data
		if len(keys) <= 0 {
			return nil
		}

		deleted, err := v.client.Unlink(ctx, keys)
		if err != nil {
			return err
		}

		v.logger.Debug(ctx, "Unlinked keys from valkey", "count", deleted)

		cursor = res.Cursor
		if cursor.IsFinished() {
			break
		}
	}
	return nil
}

func getShortUrlsByUserIdCachePrefix(userId uuid.UUID) string {
	return fmt.Sprintf("{shurl_user:id::%s}:short_urls_query", userId.String())
}

func getShortUrlsByUserIdCacheKey(userId uuid.UUID, size int, offset int) string {
	return fmt.Sprintf("%s:size::%d:offset::%d", getShortUrlsByUserIdCachePrefix(userId), size, offset)
}

func getShortUrlByIdCachePrefix(id uuid.UUID) string {
	return fmt.Sprintf("short_url:id::%s", id.String())
}

func getShortUrlBySlugCachePrefix(slug string) string {
	return fmt.Sprintf("short_url:slug::%s", slug)
}
