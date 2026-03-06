package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	HeadersIdempotencyKey       = "X-Idempotency-Key"
	HeadersContentTypeKey       = "Content-Type"
	HeadersContentTypeJsonValue = "application/json"
)

type ShortUrl struct {
	Id             uuid.UUID  `json:"id"`
	DestinationUrl string     `json:"destination_url"`
	Slug           string     `json:"slug"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	UserId         *uuid.UUID `json:"user_id,omitempty"`
}

type ShortUrlResponse struct {
	Id             *uuid.UUID `json:"id,omitempty"`
	DestinationUrl *string    `json:"destination_url,omitempty"`
	Slug           *string    `json:"slug,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Url            string     `json:"url,omitempty"`
	UserId         *uuid.UUID `json:"user_id,omitempty"`
	Errors         []string   `json:"errors,omitempty"`
}

type CreateShortUrl struct {
	Id             uuid.UUID
	DestinationUrl string
	Slug           string
	UserId         *uuid.UUID
	ExpiresAt      time.Time
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
