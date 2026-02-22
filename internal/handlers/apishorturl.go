package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
)

type ApiShortUrlHandler struct {
	Logger  utils.CustomJsonLogger
	Db      db.DbContext
	BaseUrl string
}

func NewApiShortUrlHandler(logger utils.CustomJsonLogger, dbcontext db.DbContext, baseUrl string) ApiShortUrlHandler {
	return ApiShortUrlHandler{Logger: logger, Db: dbcontext, BaseUrl: baseUrl}
}

type PostShortUrlRequest struct {
	DestinationUrl string `json:"destination_url"`
}

func (h *ApiShortUrlHandler) PostShortUrl(w http.ResponseWriter, r *http.Request) {
	var req PostShortUrlRequest

	idempotencyKeyString := r.Header.Get(types.HeadersIdempotencyKey)
	idempotencyKey, err := uuid.Parse(idempotencyKeyString)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: "idempotency key provided is not a valid UUID"})
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorCode, message := parseJsonDecodeError(err)
		EncodeResponse[types.ErrorResponse](w, errorCode, types.ErrorResponse{Error: message})
		if errorCode == http.StatusInternalServerError {
			h.Logger.Error(r.Context(), "Server error when parsing json body. error: %v", "error", err.Error())
		}
		return
	}

	req.DestinationUrl = strings.TrimSpace(req.DestinationUrl)
	if req.DestinationUrl == "" {
		EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: "`destination_url` cannot be null or empty"})
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	slug, err := h.generateUniqueSlug(r.Context())
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	newShortUrl := types.CreateShortUrl{
		Id:             id,
		DestinationUrl: req.DestinationUrl,
		Slug:           slug,
	}

	requestHash := db.HashCreateShortUrlRequest(req.DestinationUrl)
	shortUrl, err := h.Db.CreateShortUrl(r.Context(), newShortUrl, idempotencyKey, requestHash)
	if err != nil {
		var idempotencyKeyUsedError *types.DuplicateIdempotencyKeyError
		if errors.As(err, &idempotencyKeyUsedError) {
			EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("%s header value has already been used", types.HeadersIdempotencyKey)})
			h.Logger.Error(r.Context(), err.Error())
			return
		}

		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	response := types.CreateShortUrlResponse{
		Id:             shortUrl.Id,
		DestinationUrl: shortUrl.DestinationUrl,
		Slug:           shortUrl.Slug,
		CreatedAt:      shortUrl.CreatedAt,
		Url:            createShortUrl(h.BaseUrl, shortUrl.Slug),
	}

	if err = EncodeResponse[types.CreateShortUrlResponse](w, http.StatusCreated, response); err != nil {
		h.Logger.Error(r.Context(), err.Error())
	}
	h.Logger.Debug(r.Context(), "PostShortUrl created short url with id '%s'", "shortUrlId", shortUrl.Id, "responseStatusCode", 201)
}

func (h *ApiShortUrlHandler) generateUniqueSlug(ctx context.Context) (string, error) {
	slug, err := GenerateSlug()
	if err != nil {
		return "", err
	}

	existingShortUrl, err := h.Db.GetShortUrlBySlug(ctx, slug)
	if err != nil {
		return "", err
	}

	for existingShortUrl != nil {
		existingShortUrl, err = h.Db.GetShortUrlBySlug(ctx, slug)
		if err != nil {
			return "", err
		}
	}

	return slug, nil
}

func GenerateSlug() (string, error) {
	valueRange := big.NewInt(5) // Generate a random number [0, 1, 2, 3 , 4]
	n, err := rand.Int(rand.Reader, valueRange)
	if err != nil {
		return "", err
	}
	slugLength := int(n.Int64() + 4)

	possibleChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var result strings.Builder

	for i := 0; i < slugLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt((int64(len(possibleChars)))))
		if err != nil {
			return "", err
		}

		result.WriteByte(possibleChars[n.Int64()])
	}

	return result.String(), nil
}

func createShortUrl(baseUrl string, slug string) string {
	return fmt.Sprintf("%s/%s", baseUrl, slug)
}
