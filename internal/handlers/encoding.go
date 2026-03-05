package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/go-playground/validator/v10"
)

type EncodeResponseError struct{}

func (err *EncodeResponseError) Error() string {
	return "Error encoding response"
}

func EncodeResponse[T any](logger utils.CustomJsonLogger, ctx context.Context, w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set(types.HeadersContentTypeKey, types.HeadersContentTypeJsonValue)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error(ctx, "error encoding json response", "error", err.Error())
	}
}

func EncodeValidationError(err validator.ValidationErrors) []string {
	var validationErrors []string

	for _, e := range err {
		validationErrors = append(validationErrors, e.Error())
	}
	return validationErrors
}
