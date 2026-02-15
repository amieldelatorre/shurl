package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
)

func parseJsonDecodeError(err error) (int, string) {
	var invalidUnmarshalError *json.InvalidUnmarshalError
	if errors.As(err, &invalidUnmarshalError) {
		return http.StatusInternalServerError, "Something went wrong with the server, please try again later"
	}

	return http.StatusBadRequest, "Invalid JSON body for request"
}
