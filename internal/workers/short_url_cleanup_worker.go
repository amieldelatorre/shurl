package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
)

func ShortUrlCleanupWorker(ctx context.Context, logger utils.CustomJsonLogger, intervalSeconds int, dbContext db.DbContext, errorsFatal bool) {
	ctx = context.WithValue(ctx, utils.RequestIdName, "shortUrlCleanupWorker")
	logger.Info(ctx, fmt.Sprintf("starting short url cleanup worker with interval an of %d seconds", intervalSeconds))
	handlers.ShortUrlCleanupWorkerRunning = true

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "signal received, shutting down short url cleanup worker")
			handlers.ShortUrlCleanupWorkerRunning = false
			return
		case <-ticker.C:
			logger.Debug(ctx, "short url cleanup worker woken up, performing cleanup")
			err := performShortUrlCleanup(ctx, logger, dbContext)
			if err != nil {
				logger.Error(ctx, err.Error())
				if errorsFatal {
					logger.Error(ctx, "short_url_cleanup_worker.errors_fatal is set to true, exiting worker")
					handlers.ShortUrlCleanupWorkerRunning = false
					return
				}
			}

			logger.Debug(ctx, fmt.Sprintf("short url cleanup worker sleeping for %d seconds", intervalSeconds))
		}
	}
}

func performShortUrlCleanup(ctx context.Context, logger utils.CustomJsonLogger, dbContext db.DbContext) error {
	numCleaned, err := dbContext.DeleteExpiredShortUrls(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Number of short urls cleaned: %d", numCleaned))
	return nil
}
