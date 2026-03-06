package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	// m=9216 (9 MiB), t=4, p=1
	argon2idParams = &argon2id.Params{
		Memory:      256 * 1024,
		Iterations:  4,
		Parallelism: 2, // requires 2 cpu cores
		SaltLength:  16,
		KeyLength:   32,
	}
)

type ApiUserHandler struct {
	Logger utils.CustomJsonLogger
	Db     db.DbContext
}

func NewApiUserHandler(logger utils.CustomJsonLogger, dbContext db.DbContext) ApiUserHandler {
	return ApiUserHandler{Logger: logger, Db: dbContext}
}

type PostUserRequest struct {
	Username        string `json:"username" validate:"required,alphanum,lowercase,min=3"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type PostUserResponse struct {
	Id        *uuid.UUID `json:"id,omitempty"`
	Username  *string    `json:"username,omitempty"`
	Email     *string    `json:"email,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Errors    []string   `json:"errors,omitempty"`
}

func (h *ApiUserHandler) PostUser(w http.ResponseWriter, r *http.Request) {
	var req PostUserRequest

	// get idempotency key from header
	idempotencyKeyString := r.Header.Get(types.HeadersIdempotencyKey)
	idempotencyKey, err := uuid.Parse(idempotencyKeyString)
	if err != nil {
		EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusBadRequest, PostUserResponse{Errors: []string{"idempotency key provided is not a valid UUID"}})
		return
	}

	// decode body
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorCode, message := parseJsonDecodeError(err)
		EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusBadRequest, PostUserResponse{Errors: []string{message}})
		if errorCode == http.StatusInternalServerError {
			h.Logger.Error(r.Context(), "Server error when parsing json body. error: %v", "error", err.Error())
		}
		return
	}

	newUser, err := CreateUser(r.Context(), h.Db, req, idempotencyKey)
	if err != nil {
		var validationError validator.ValidationErrors
		var duplicateIdempotencyKeyError *types.DuplicateIdempotencyKeyError
		var duplicateEmailOrUsername *types.EmailOrUsernameExistsError

		switch {
		case errors.As(err, &validationError):
			EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusBadRequest, PostUserResponse{Errors: EncodeValidationError(validationError)})
		case errors.As(err, &duplicateIdempotencyKeyError):
			EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusBadRequest, PostUserResponse{Errors: []string{fmt.Sprintf("%s header value has already been used", types.HeadersIdempotencyKey)}})
		case errors.As(err, &duplicateEmailOrUsername):
			EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusBadRequest, PostUserResponse{Errors: []string{err.Error()}})
		default:
			EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, PostUserResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
			h.Logger.Error(r.Context(), err.Error())
		}
		return
	}

	response := PostUserResponse{
		Id:        &newUser.Id,
		Username:  &newUser.Username,
		Email:     &newUser.Email,
		CreatedAt: &newUser.CreatedAt,
		UpdatedAt: &newUser.UpdatedAt,
	}

	EncodeResponse[PostUserResponse](h.Logger, r.Context(), w, http.StatusCreated, response)
	h.Logger.Info(r.Context(), "PostUser created user with id '%s'", "userId", newUser.Id, "responseStatusCode", 201)
}

func CreateUser(ctx context.Context, dbContext db.DbContext, requestedUser PostUserRequest, idempotencyKey uuid.UUID) (*types.User, error) {
	// validate request
	validate, err := utils.GetValidator()
	if err != nil {
		return nil, err
	}

	err = validate.Struct(&requestedUser)
	if err != nil {
		return nil, err
	}

	// create password hash
	hashedPassword, err := argon2id.CreateHash(requestedUser.Password, argon2idParams)
	if err != nil {
		return nil, err
	}

	// create new user
	requestHash := db.HashCreateUserRequest(requestedUser.Username, requestedUser.Email)
	newUserId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	newUser, err := dbContext.CreateUser(ctx, idempotencyKey, requestHash, types.CreateUserRequest{
		Id:           newUserId,
		Username:     requestedUser.Username,
		Email:        requestedUser.Email,
		PasswordHash: hashedPassword,
	})
	return newUser, err
}
