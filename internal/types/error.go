package types

type ErrorResponse struct {
	Error string `json:"error"`
}

type ErrorResponseList struct {
	Errors []string `json:"errors"`
}

type DuplicateIdempotencyKeyError struct {
}

func (e *DuplicateIdempotencyKeyError) Error() string {
	return "Idempotency key already used"
}

type EmailOrUsernameExistsError struct{}

func (e *EmailOrUsernameExistsError) Error() string {
	return "Username or email already exists"
}
