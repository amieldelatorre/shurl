package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
)

type Middlware struct {
	Logger utils.CustomJsonLogger
	Config config.Config
}

func NewMiddleware(logger utils.CustomJsonLogger, config config.Config) Middlware {
	return Middlware{Logger: logger, Config: config}
}

func (m *Middlware) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.Logger.Error(r.Context(), "Had to recover from panic", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (m *Middlware) AddRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Logger.Debug(r.Context(), "Adding a request id to incoming request")
		id, err := uuid.NewV7()
		if err != nil {
			m.Logger.Error(r.Context(), "Problem generating uuid", "error", err)
		}

		m.Logger.Debug(r.Context(), "Request id generated and added to context", string(utils.RequestIdName), id.String())
		ctx := context.WithValue(r.Context(), utils.RequestIdName, id.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middlware) IdempotencyKeyRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idempotencyKey := r.Header.Get(types.HeadersIdempotencyKey)
		if strings.TrimSpace(idempotencyKey) == "" {
			EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("Missing uuidv7 idempotency key header '%s'", types.HeadersIdempotencyKey)})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middlware) AllowRegistration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Config.Server.AllowRegistration && m.Config.Server.AllowLogin {
			next.ServeHTTP(w, r)
			return
		}

		EncodeResponse[types.ErrorResponse](w, http.StatusForbidden, types.ErrorResponse{Error: "Registration has been disabled by the administrator"})
	})
}
