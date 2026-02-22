package types

import (
	"time"

	"github.com/google/uuid"
)

const HeadersIdempotencyKey = "X-Idempotency-Key"

type ShortUrl struct {
	Id             uuid.UUID `json:"id"`
	DestinationUrl string    `json:"destination_url"`
	Slug           string    `json:"slug"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateShortUrlResponse struct {
	Id             uuid.UUID `json:"id"`
	DestinationUrl string    `json:"destination_url"`
	Slug           string    `json:"slug"`
	CreatedAt      time.Time `json:"created_at"`
	Url            string    `json:"url"`
}

type CreateShortUrl struct {
	Id             uuid.UUID
	DestinationUrl string
	Slug           string
}

type CreateUserRequest struct {
	Id           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
}

type User struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
