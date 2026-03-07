package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	SizeQueryParamError                    = "Invalid page value, must be a number greater than or equal to 1 and less than or equal to 50"
	DefaultSizeQueryParam                  = "20"
	PageQueryParamError                    = "Invalid page value, must be a number greater than or equal to 1"
	DefaultPageQueryParam                  = "1"
	DefaultAnonymousShortUrlTtl     uint32 = 259200 // 3 days
	MaxAnonymousShortUrlTtl         uint32 = 604800 // 7 Days
	DefaultAuthenticatedShortUrlTtl uint32 = 604800 // 7 days
	MaxAuthenticatedShortUrlTtl
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
	DestinationUrl string  `json:"destination_url" validate:"required,url"`
	TTL            *uint32 `json:"ttl" validate:"required,min=900,max=2629746"` // 15 minutes to 1 months
}

func (h *ApiShortUrlHandler) PostShortUrl(w http.ResponseWriter, r *http.Request) {
	var req PostShortUrlRequest

	userIdValue := r.Context().Value(UserIdKey)
	userIdUuid, ok := userIdValue.(uuid.UUID)
	if !ok {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), "casting uuid from context not ok")
		return
	}

	idempotencyKeyString := r.Header.Get(types.HeadersIdempotencyKey)
	idempotencyKey, err := uuid.Parse(idempotencyKeyString)
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusBadRequest, types.ErrorResponse{Errors: []string{"idempotency key provided is not a valid UUID"}})
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorCode, message := parseJsonDecodeError(err)
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, errorCode, types.ErrorResponse{Errors: []string{message}})
		if errorCode == http.StatusInternalServerError {
			h.Logger.Error(r.Context(), "Server error when parsing json body. error: %v", "error", err.Error())
		}
		return
	}

	if req.TTL != nil {
		// anonymous users can only have a maximum of 7 days
		if userIdUuid == uuid.Nil && *req.TTL > MaxAnonymousShortUrlTtl {
			EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusBadRequest, types.ErrorResponse{Errors: []string{fmt.Sprintf("anonymous short urls can only be up to %d seconds", MaxAnonymousShortUrlTtl)}})
			return
		}
	} else {
		var ttl uint32
		if userIdUuid == uuid.Nil {
			ttl = DefaultAnonymousShortUrlTtl
		} else {
			ttl = DefaultAuthenticatedShortUrlTtl
		}
		req.TTL = &ttl
	}

	req.DestinationUrl = strings.TrimSpace(req.DestinationUrl)
	validate, err := utils.GetValidator()
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}
	var validationError validator.ValidationErrors
	err = validate.Struct(&req)
	if err != nil {
		if errors.As(err, &validationError) {
			EncodeResponse[types.ShortUrlResponse](h.Logger, r.Context(), w, http.StatusBadRequest, types.ShortUrlResponse{Errors: EncodeValidationError(validationError)})
			return
		}
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}
	// if req.DestinationUrl == "" {
	// 	EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusBadRequest, types.ErrorResponse{Errors: []string{"`destination_url` cannot be null or empty"}})
	// 	return
	// }

	id, err := uuid.NewV7()
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	slug, err := h.generateUniqueSlug(r.Context())
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	newShortUrl := types.CreateShortUrl{
		Id:             id,
		DestinationUrl: req.DestinationUrl,
		Slug:           slug,
		ExpiresAt:      time.Now().Add(time.Duration(*req.TTL) * time.Second),
	}
	if userIdUuid != uuid.Nil {
		newShortUrl.UserId = &userIdUuid
	}

	requestHash := db.HashCreateShortUrlRequest(req.DestinationUrl)
	shortUrl, err := h.Db.CreateShortUrl(r.Context(), newShortUrl, idempotencyKey, requestHash)
	if err != nil {
		var idempotencyKeyUsedError *types.DuplicateIdempotencyKeyError
		if errors.As(err, &idempotencyKeyUsedError) {
			EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusBadRequest, types.ErrorResponse{Errors: []string{fmt.Sprintf("%s header value has already been used", types.HeadersIdempotencyKey)}})
			h.Logger.Error(r.Context(), err.Error())
			return
		}

		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	response := types.ShortUrlResponse{
		Id:             &shortUrl.Id,
		DestinationUrl: &shortUrl.DestinationUrl,
		Slug:           &shortUrl.Slug,
		CreatedAt:      &shortUrl.CreatedAt,
		ExpiresAt:      &shortUrl.ExpiresAt,
		Url:            createShortUrl(h.BaseUrl, shortUrl.Slug),
		UserId:         shortUrl.UserId,
	}

	EncodeResponse[types.ShortUrlResponse](h.Logger, r.Context(), w, http.StatusCreated, response)
	h.Logger.Debug(r.Context(), "PostShortUrl created short url with id '%s'", "shortUrlId", shortUrl.Id, "responseStatusCode", 201)
}

func (h *ApiShortUrlHandler) generateUniqueSlug(ctx context.Context) (string, error) {
	maxAttempts := 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		slug, err := GenerateSlug()
		if err != nil {
			return "", err
		}
		existingShortUrl, err := h.Db.GetShortUrlBySlug(ctx, slug, false)
		if err != nil {
			return "", err
		}
		if existingShortUrl == nil {
			return slug, nil
		}
	}
	return "", errors.New("couldn't generate a unique slug")
}

type GetShortUrlsByUserIdResponse struct {
	Items  []types.ShortUrlResponse `json:"items"`
	Total  *int                     `json:"total,omitempty"`
	Next   *bool                    `json:"next,omitempty"`
	Page   *int                     `json:"page,omitempty"`
	Size   *int                     `json:"size,omitempty"`
	Errors []string                 `json:"errors,omitempty"`
}

func (h *ApiShortUrlHandler) GetShortUrls(w http.ResponseWriter, r *http.Request) {
	userIdValue := r.Context().Value(UserIdKey)
	userIdUuid, ok := userIdValue.(uuid.UUID)
	if !ok {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), "casting uuid from context not ok")
		return
	}

	params := r.URL.Query()

	pageStr := strings.TrimSpace(params.Get("page"))
	if pageStr == "" {
		pageStr = DefaultPageQueryParam
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		EncodeResponse[GetShortUrlsByUserIdResponse](h.Logger, r.Context(), w, http.StatusBadRequest, GetShortUrlsByUserIdResponse{Errors: []string{PageQueryParamError}})
		return
	}

	sizeStr := strings.TrimSpace(params.Get("size"))
	if sizeStr == "" {
		sizeStr = DefaultSizeQueryParam
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 50 {
		EncodeResponse[GetShortUrlsByUserIdResponse](h.Logger, r.Context(), w, http.StatusBadRequest, GetShortUrlsByUserIdResponse{Errors: []string{SizeQueryParamError}})
		return
	}

	// Subtract 1 from offset because this actually does 0 indexing
	// For a users persective a page 0 doesn't really exist, page 1 is where they expect to see the first items
	offset := (page - 1) * size
	shortUrls, err := h.Db.GetShortUrlsByUserId(r.Context(), userIdUuid, size, offset)
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	resp := shortUrlToResponse(shortUrls, h.BaseUrl, page, size)
	EncodeResponse[GetShortUrlsByUserIdResponse](h.Logger, r.Context(), w, http.StatusOK, resp)
}

func (h *ApiShortUrlHandler) DeleteById(w http.ResponseWriter, r *http.Request) {
	userIdValue := r.Context().Value(UserIdKey)
	userIdUuid, ok := userIdValue.(uuid.UUID)
	if !ok {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), "casting uuid from context not ok")
		return
	}

	shortUrlIdStr := strings.TrimSpace(r.PathValue("shortUrlId"))
	shortUrlid, err := uuid.Parse(shortUrlIdStr)
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusBadRequest, types.ErrorResponse{Errors: []string{"Short url id provided is not a valid uuid"}})
		return
	}

	delRes, err := h.Db.DeleteShortUrlById(r.Context(), userIdUuid, shortUrlid)
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	if !delRes.Found && delRes.NumDeleted == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if delRes.Found && delRes.NumDeleted == 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.Logger.Error(r.Context(), "reached end of short url delete by id. this should not happen")
}

func shortUrlToResponse(shortUrls types.GetShortUrlsResult, baseUrl string, page int, size int) GetShortUrlsByUserIdResponse {
	resp := GetShortUrlsByUserIdResponse{
		Items: []types.ShortUrlResponse{},
		Total: &shortUrls.Total,
		Page:  &page,
		Size:  &size,
	}

	var p int
	if page == 0 {
		p = 1
	} else {
		p = page
	}
	next := p*size < shortUrls.Total
	resp.Next = &next

	for _, s := range shortUrls.Items {
		r := types.ShortUrlResponse{
			Id:             &s.Id,
			DestinationUrl: &s.DestinationUrl,
			Slug:           &s.Slug,
			CreatedAt:      &s.CreatedAt,
			ExpiresAt:      &s.ExpiresAt,
			Url:            createShortUrl(baseUrl, s.Slug),
			UserId:         s.UserId,
			Errors:         []string{},
		}

		resp.Items = append(resp.Items, r)
	}
	return resp
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
