package handlers

import (
	"fmt"
	"net/http"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/utils"
)

var (
	IdempotencyKeyCleanupWorkerRunning = false
	ShortUrlCleanupWorkerRunning       = false
)

type ApiHealthHandler struct {
	Logger utils.CustomJsonLogger
	Config config.Config
	Db     db.DbContext
	Cache  db.DbContext
}

func NewApiHealthHandler(logger utils.CustomJsonLogger, config config.Config, dbContext db.DbContext, cache db.DbContext) ApiHealthHandler {
	return ApiHealthHandler{
		Logger: logger,
		Config: config,
		Db:     dbContext,
		Cache:  cache,
	}
}

type HealthCheckResponse struct {
	IdempotencyKeyCleanupWorker IdempotencyKeyCleanupWorkerHealthCheck `json:"idempotency_key_cleanup_worker"`
	ShortUrlCleanUpWorker       ShortUrlCleanupWorkerHealthCheck       `json:"short_url_cleanup_worker"`
	Database                    DatabaseHealthCheck                    `json:"database"`
	Cache                       CacheHealthCheck                       `json:"cache"`
	Errors                      []string                               `json:"errors,omitempty"`
}

type DatabaseHealthCheck struct {
	Ok      bool   `json:"ok"`
	Version string `json:"version,omitempty"`
}

type CacheHealthCheck struct {
	Enabled bool  `json:"enabled"`
	Ok      *bool `json:"ok,omitempty"`
}

type IdempotencyKeyCleanupWorkerHealthCheck struct {
	Running bool `json:"running"`
}

type ShortUrlCleanupWorkerHealthCheck struct {
	Running bool `json:"running"`
}

func (h *ApiHealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	errs := []string{}
	status := http.StatusOK

	response := HealthCheckResponse{
		IdempotencyKeyCleanupWorker: IdempotencyKeyCleanupWorkerHealthCheck{
			Running: IdempotencyKeyCleanupWorkerRunning,
		},
		ShortUrlCleanUpWorker: ShortUrlCleanupWorkerHealthCheck{
			Running: ShortUrlCleanupWorkerRunning,
		},
		Database: DatabaseHealthCheck{
			Ok: true,
		},
		Cache: CacheHealthCheck{
			Enabled: *h.Config.Cache.Enabled,
		},
	}

	err := h.Db.Ping(r.Context())
	if err != nil {
		errs = append(errs, "could not ping database")
		response.Database = DatabaseHealthCheck{
			Ok: false,
		}
		h.Logger.Error(r.Context(), err.Error())
	}

	version, err := h.Db.GetDatabaseVersion(r.Context())
	if err != nil {
		errs = append(errs, "could not get db version")
		response.Database = DatabaseHealthCheck{
			Ok: false,
		}
		h.Logger.Error(r.Context(), err.Error())
	}

	response.Database.Version = fmt.Sprintf("%d", version)

	if *h.Config.Cache.Enabled {
		err = h.Cache.Ping(r.Context())

		if err != nil {
			h.Logger.Error(r.Context(), err.Error())
			ok := false
			response.Cache.Ok = &ok
			errs = append(errs, "could not ping cache")
		} else {
			ok := true
			response.Cache.Ok = &ok
		}
	}

	response.Errors = errs
	if len(errs) > 0 {
		status = http.StatusInternalServerError
	}
	EncodeResponse[HealthCheckResponse](h.Logger, r.Context(), w, status, response)
}
