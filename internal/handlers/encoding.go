package handlers

import (
	"encoding/json"
	"net/http"
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
