package handlers

import (
	"net/http"
	"strings"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
)

type RedirectionHandler struct {
	Logger utils.CustomJsonLogger
	Db     db.DbContext
}

func NewRedirectionHandler(logger utils.CustomJsonLogger, db db.DbContext) RedirectionHandler {
	return RedirectionHandler{Logger: logger, Db: db}
}

func (h *RedirectionHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if len(slug) < 4 {
		http.NotFound(w, r)
		return
	}

	destination, err := h.Db.GetShortUrlBySlug(r.Context(), slug)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	if destination == nil {
		http.NotFound(w, r)
		h.Logger.Debug(r.Context(), "Unknown slug", "slug", slug)
		return
	}

	http.Redirect(w, r, destination.DestinationUrl, http.StatusTemporaryRedirect)
	h.Logger.Info(r.Context(), "Redirect", "responseStatusCode", http.StatusTemporaryRedirect)
}
