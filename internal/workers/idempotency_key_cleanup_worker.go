package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
)

func IdempotencyKeyCleanupWorker(ctx context.Context, logger utils.CustomJsonLogger, intervalSeconds int, dbContext db.DbContext, errorsFatal bool) {
	ctx = context.WithValue(ctx, utils.RequestIdName, "idempotencyKeyCleanupWorker")
	logger.Info(ctx, "starting idempotency key cleanup worker")
	handlers.IdempotencyKeyCleanupWorkerRunning = true

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "signal received, shutting down idempotency key cleanup worker")
			handlers.IdempotencyKeyCleanupWorkerRunning = false
			return
		case <-ticker.C:
			logger.Info(ctx, "idempotency key cleanup worker woken up, performing cleanup")
			err := performIdempotencyKeyCleanup(ctx, logger, dbContext)
			if err != nil {
				logger.Error(ctx, err.Error())
				if errorsFatal {
					logger.Error(ctx, "idempotency_key_cleanup_worker.errors_fatal is set to true, exiting worker")
					handlers.IdempotencyKeyCleanupWorkerRunning = false
					return
				}
			}
		}
	}
}

func performIdempotencyKeyCleanup(ctx context.Context, logger utils.CustomJsonLogger, dbContext db.DbContext) error {
	numCleaned, err := dbContext.DeleteExpiredIdempotencyKeys(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Number of idempotency keys cleaned: %d", numCleaned))
	return nil
}
