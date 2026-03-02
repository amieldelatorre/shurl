package handlers

import (
	"fmt"
	"net/http"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/utils"
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
	Database DatabaseHealthCheck `json:"database"`
}

type DatabaseHealthCheck struct {
	Ok      bool   `json:"ok"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *ApiHealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.Db.Ping(r.Context())
	if err != nil {
		EncodeResponse[HealthCheckResponse](w, http.StatusInternalServerError,
			HealthCheckResponse{
				Database: DatabaseHealthCheck{
					Ok:    false,
					Error: "could not ping database",
				},
			})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	version, err := h.Db.GetDatabaseVersion(r.Context())
	if err != nil {
		EncodeResponse[HealthCheckResponse](w, http.StatusInternalServerError,
			HealthCheckResponse{
				Database: DatabaseHealthCheck{
					Ok:    false,
					Error: "could not get db version",
				},
			})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	EncodeResponse[HealthCheckResponse](w, http.StatusOK, HealthCheckResponse{
		Database: DatabaseHealthCheck{
			Ok:      true,
			Version: fmt.Sprintf("%d", version),
		},
	})
}
