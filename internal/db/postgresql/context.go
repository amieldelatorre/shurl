package postgresql

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQLContext struct {
	logger utils.CustomJsonLogger
	dbPool *pgxpool.Pool
}

func NewPostreSQLContext(logger utils.CustomJsonLogger, dbPool *pgxpool.Pool) *PostgreSQLContext {
	return &PostgreSQLContext{logger: logger, dbPool: dbPool}
}

func (p *PostgreSQLContext) GetDatabaseVersion() int64 {
	return 2
}

func (p *PostgreSQLContext) GetShortUrlById(ctx context.Context, id uuid.UUID) (*types.ShortUrl, error) {
	return ExecWithRetry(ctx, p.logger, p.dbPool, func(tx pgx.Tx) (*types.ShortUrl, error) {
		return p.getShortUrlByIdWithTx(ctx, tx, id)
	})
}

func (p *PostgreSQLContext) getShortUrlByIdWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*types.ShortUrl, error) {
	var shortUrl types.ShortUrl

	// slug should be unique
	err := tx.QueryRow(ctx, `SELECT id, destination_url, slug, created_at FROM short_urls WHERE id = $1`, id).Scan(
		&shortUrl.Id, &shortUrl.DestinationUrl, &shortUrl.Slug, &shortUrl.CreatedAt,
	)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &shortUrl, err
}

func (p *PostgreSQLContext) CreateShortUrl(ctx context.Context, req types.CreateShortUrl, idempotencyKey uuid.UUID) (*types.ShortUrl, error) {
	return ExecWithRetry(ctx, p.logger, p.dbPool, func(tx pgx.Tx) (*types.ShortUrl, error) {
		var newShortUrl types.ShortUrl
		idempotencyKeyUuid, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}

		var idempotencyKeyReferenceId uuid.UUID
		err = tx.QueryRow(ctx,
			`INSERT INTO idempotency_keys (id, i_key, reference_id, created_at, expires_at)
			 VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '24 hours')
			 ON CONFLICT (i_key) DO UPDATE set id = EXCLUDED.i_key
			 RETURNING reference_id`, // EXCLUDED is a virtual table that that has the values we just tried to insert but couldn't due to the conflict
			idempotencyKeyUuid, idempotencyKey, req.Id).Scan(&idempotencyKeyReferenceId)
		if err != nil {
			return nil, err
		}

		// if the reference id doesn't match the current request's id, send the old data back
		if idempotencyKeyReferenceId != req.Id {
			return p.getShortUrlByIdWithTx(ctx, tx, idempotencyKeyReferenceId)
		}

		err = tx.QueryRow(ctx,
			`INSERT INTO short_urls (id, destination_url, slug, created_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (id) DO UPDATE set id = EXCLUDED.id
			 RETURNING id, destination_url, slug, created_at`,
			req.Id, req.DestinationUrl, req.Slug).Scan(
			&newShortUrl.Id, &newShortUrl.DestinationUrl, &newShortUrl.Slug, &newShortUrl.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		return &newShortUrl, nil
	})
}

func (p *PostgreSQLContext) GetShortUrlBySlug(ctx context.Context, slug string) (*types.ShortUrl, error) {
	return ExecWithRetry(ctx, p.logger, p.dbPool, func(tx pgx.Tx) (*types.ShortUrl, error) {
		var shortUrl types.ShortUrl

		// slug should be unique
		err := tx.QueryRow(ctx, `SELECT id, destination_url, slug, created_at FROM short_urls WHERE slug = $1`, slug).Scan(
			&shortUrl.Id, &shortUrl.DestinationUrl, &shortUrl.Slug, &shortUrl.CreatedAt,
		)
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return &shortUrl, err
	})
}

func ExecWithRetry[T any](ctx context.Context, logger utils.CustomJsonLogger, dbPool *pgxpool.Pool, fn func(pgx.Tx) (T, error)) (T, error) {
	var noResult T
	maxAttempts := 3
	initialDelay := 150 * time.Millisecond
	for attempt := 0; attempt < maxAttempts; attempt++ {
		res, err := func() (T, error) {
			var noResult T
			tx, err := dbPool.Begin(ctx)
			if err != nil {
				return noResult, err
			}
			// use a different context here so that even if it does time out it still runs the rollback
			defer tx.Rollback(context.Background())

			// Perform the query
			val, err := fn(tx)
			if err != nil {
				return noResult, err
			}

			if err = tx.Commit(ctx); err != nil {
				return noResult, err
			}
			return val, nil
		}()

		if err == nil {
			return res, nil
		}

		if !isRetryable(err) || attempt == maxAttempts-1 {
			return noResult, err
		}

		logger.Warn(ctx, "retryable error detected, trying again", "err", err.Error())

		// delay and jitter to prevent overwhelming the servers
		delay := time.Duration(float64(initialDelay) * math.Pow(2, float64(attempt)))
		jitter := time.Duration(rand.Int63n(int64(delay / 4)))
		select {
		case <-time.After(delay + jitter):
		case <-ctx.Done():
			return noResult, ctx.Err()
		}
	}

	return noResult, context.DeadlineExceeded
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var pgErrType *pgconn.PgError
	if errors.As(err, &pgErrType) {
		switch pgErrType.Code {
		case "40001", // serialization_failure
			"40P01", // deadlock_detected
			"53300", // too_many_connections
			"53000", //	insufficient_resources
			"53100", // disk_full
			"53200", // out_of_memory
			"08001", //	sqlclient_unable_to_establish_sqlconnection
			"08004", //	sqlserver_rejected_establishment_of_sqlconnection
			"08006", // connection_failure
			"57P01", // admin_shutdown
			"57P02", // crash_shutdown
			"57P03": // cannot_connect_now
			return true
		}
	}

	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	var netErrType net.Error
	if errors.As(err, &netErrType) {
		return netErrType.Timeout()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}
