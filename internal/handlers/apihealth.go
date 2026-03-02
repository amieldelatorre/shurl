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
)

type ApiHealthHandler struct {
	Logger utils.CustomJsonLogger
	Config config.Config
	Db     db.DbContext
}

func NewApiHealthHandler(logger utils.CustomJsonLogger, config config.Config, dbContext db.DbContext) ApiHealthHandler {
	return ApiHealthHandler{
		Logger: logger,
		Config: config,
		Db:     dbContext,
	}
}

type HealthCheckResponse struct {
	IdempotencyKeyCleanupWorker IdempotencyKeyCleanupWorkerHealthCheck `json:"idempotency_key_cleanup_worker"`
	Database                    DatabaseHealthCheck                    `json:"database"`
}

type DatabaseHealthCheck struct {
	Ok      bool   `json:"ok"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

type IdempotencyKeyCleanupWorkerHealthCheck struct {
	Running bool `json:"running"`
}

func (h *ApiHealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthCheckResponse{
		IdempotencyKeyCleanupWorker: IdempotencyKeyCleanupWorkerHealthCheck{
			Running: IdempotencyKeyCleanupWorkerRunning,
		}}
	err := h.Db.Ping(r.Context())
	if err != nil {
		response.Database = DatabaseHealthCheck{
			Ok:    false,
			Error: "could not ping database",
		}
		EncodeResponse[HealthCheckResponse](w, http.StatusInternalServerError, response)
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	version, err := h.Db.GetDatabaseVersion(r.Context())
	if err != nil {
		response.Database = DatabaseHealthCheck{
			Ok:    false,
			Error: "could not get db version",
		}
		EncodeResponse[HealthCheckResponse](w, http.StatusInternalServerError, response)
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	response.Database = DatabaseHealthCheck{
		Ok:      true,
		Version: fmt.Sprintf("%d", version),
	}
	EncodeResponse[HealthCheckResponse](w, http.StatusOK, response)
}
