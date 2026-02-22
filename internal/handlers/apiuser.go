package handlers

import (
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

type ApiUserHandler struct {
	Logger utils.CustomJsonLogger
	Db     db.DbContext
}

func NewApiUserHandler(logger utils.CustomJsonLogger, dbContext db.DbContext) ApiUserHandler {
	return ApiUserHandler{Logger: logger, Db: dbContext}
}

type postUserRequest struct {
	Username        string `json:"username" validate:"required,alphanum,lowercase,min=3"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type postUserResponse struct {
	Id        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *ApiUserHandler) PostUser(w http.ResponseWriter, r *http.Request) {
	var req postUserRequest

	// get idempotency key from header
	idempotencyKeyString := r.Header.Get(types.HeadersIdempotencyKey)
	idempotencyKey, err := uuid.Parse(idempotencyKeyString)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: "idempotency key provided is not a valid UUID"})
		return
	}

	// decode body
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorCode, message := parseJsonDecodeError(err)
		EncodeResponse[types.ErrorResponse](w, errorCode, types.ErrorResponse{Error: message})
		if errorCode == http.StatusInternalServerError {
			h.Logger.Error(r.Context(), "Server error when parsing json body. error: %v", "error", err.Error())
		}
		return
	}

	// validate request
	validate := validator.New()
	err = validate.Struct(&req)
	if err != nil {
		var vError validator.ValidationErrors
		if errors.As(err, &vError) {
			EncodeResponse[types.ErrorResponseList](w, http.StatusBadRequest, types.ErrorResponseList{Error: EncodeValidationError(vError)})
			return
		}

		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	// create password hash
	// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	// m=9216 (9 MiB), t=4, p=1
	argon2idParams := &argon2id.Params{
		Memory:      256 * 1024,
		Iterations:  4,
		Parallelism: 2, // requires 2 cpu cores
		SaltLength:  16,
		KeyLength:   32,
	}
	hashedPassword, err := argon2id.CreateHash(req.Password, argon2idParams)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	// create new user
	requestHash := db.HashCreateUserRequest(req.Username, req.Email)
	newUserId, err := uuid.NewV7()
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	newUser, err := h.Db.CreateUser(r.Context(), idempotencyKey, requestHash, types.CreateUserRequest{
		Id:           newUserId,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		var idempotencyKeyUsedError *types.DuplicateIdempotencyKeyError
		if errors.As(err, &idempotencyKeyUsedError) {
			EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("%s header value has already been used", types.HeadersIdempotencyKey)})
			return
		}

		var uniqueViolationError *types.EmailOrUsernameExistsError
		if errors.As(err, &uniqueViolationError) {
			EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
			return
		}

		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	response := postUserResponse{
		Id:        newUser.Id,
		Username:  newUser.Username,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}

	if err = EncodeResponse[postUserResponse](w, http.StatusCreated, response); err != nil {
		h.Logger.Error(r.Context(), err.Error())
	}
	h.Logger.Info(r.Context(), "PostUser created user with id '%s'", "userId", newUser.Id, "responseStatusCode", 201)
}
