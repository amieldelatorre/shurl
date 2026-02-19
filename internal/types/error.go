package types

type ErrorResponse struct {
	Error string `json:"error"`
}

type DuplicateIdempotencyKeyError struct {
}

func (e *DuplicateIdempotencyKeyError) Error() string {
	return "Idempotency key already used"
}
