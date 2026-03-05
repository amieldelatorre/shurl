package types

type ErrorResponse struct {
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
