package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type EncodeResponseError struct{}

func (err *EncodeResponseError) Error() string {
	return "Error encoding response"
}

func EncodeResponse[T any](w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return &EncodeResponseError{}
	}

	return nil
}

func EncodeValidationError(err validator.ValidationErrors) []string {
	var validationErrors []string

	for _, e := range err {
		validationErrors = append(validationErrors, e.Error())
	}
	return validationErrors
}
