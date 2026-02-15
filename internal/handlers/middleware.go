package handlers

import (
	"context"
	"net/http"

	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
)

type Middlware struct {
	Logger utils.CustomJsonLogger
}

func NewMiddleware(logger utils.CustomJsonLogger) Middlware {
	return Middlware{Logger: logger}
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
