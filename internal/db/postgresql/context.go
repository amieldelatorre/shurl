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

func (p *PostgreSQLContext) CreateShortUrl(ctx context.Context, req types.CreateShortUrl) (*types.ShortUrl, error) {
	return ExecWithRetry(ctx, p.logger, p.dbPool, func(tx pgx.Tx) (*types.ShortUrl, error) {
		var newShortUrl types.ShortUrl
		err := p.dbPool.QueryRow(ctx,
			`INSERT INTO short_urls (id, destination_url, slug, created_at)
			 VALUES ($1, $2, $3, NOW())
			 RETURNING id, destination_url, slug, created_at`,
			req.Id, req.DestinationUrl, req.Slug).Scan(
			&newShortUrl.Id, &newShortUrl.DestinationUrl, &newShortUrl.Slug, &newShortUrl.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		return &newShortUrl, err
	})
}

func (p *PostgreSQLContext) GetShortUrlBySlug(ctx context.Context, slug string) (*types.ShortUrl, error) {
	return ExecWithRetry(ctx, p.logger, p.dbPool, func(tx pgx.Tx) (*types.ShortUrl, error) {
		var shortUrl types.ShortUrl

		// slug should be unique
		err := p.dbPool.QueryRow(ctx, `SELECT * FROM short_urls WHERE slug = $1`, slug).Scan(
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
	initialDelay := 5000 * time.Millisecond
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
