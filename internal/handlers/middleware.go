package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextKey string

const (
	UserIdKey                 ContextKey = "user_id"
	HeaderAuthorization       string     = "Authorization"
	HeaderAuthorizationPrefix string     = "Bearer "
)

var (
	CookieAccessTokenName string = "access_token"
)

type Middleware struct {
	Logger utils.CustomJsonLogger
	Config config.Config
}

func NewMiddleware(logger utils.CustomJsonLogger, config config.Config) Middleware {
	return Middleware{Logger: logger, Config: config}
}

func (m *Middleware) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.Logger.Error(r.Context(), "Had to recover from panic", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) AddRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Logger.Debug(r.Context(), "Adding a request id to incoming request")
		id, err := uuid.NewV7()
		if err != nil {
			m.Logger.Error(r.Context(), "Problem generating uuid", "error", err)
		}

		m.Logger.Debug(r.Context(), "Request id generated and added to context", string(utils.RequestIdName), id.String())
		ctx := context.WithValue(r.Context(), utils.RequestIdName, id.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) IdempotencyKeyRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idempotencyKey := r.Header.Get(types.HeadersIdempotencyKey)
		if strings.TrimSpace(idempotencyKey) == "" {
			EncodeResponse[types.ErrorResponse](w, http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("Missing uuidv7 idempotency key header '%s'", types.HeadersIdempotencyKey)})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) AllowRegistration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Config.Server.AllowRegistration && m.Config.Server.AllowLogin {
			next.ServeHTTP(w, r)
			return
		}

		EncodeResponse[types.ErrorResponse](w, http.StatusForbidden, types.ErrorResponse{Error: "Signup has been disabled by the administrator"})
	})
}

func (m *Middleware) AllowLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Config.Server.AllowLogin {
			next.ServeHTTP(w, r)
			return
		}

		EncodeResponse[types.ErrorResponse](w, http.StatusForbidden, types.ErrorResponse{Error: "Login has been disabled by the administrator"})
	})
}

func (m *Middleware) LoginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := m.GetAccessToken(r)
		if err != nil {
			EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
			m.Logger.Error(r.Context(), err.Error())
			return
		}

		if accessToken == "" {
			EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: "Login required"})
			return
		}

		claims, isValidAccessToken, err := ValidateAccessToken(accessToken, &m.Config.Server.Auth.JwtEcdsaParsedKey.PublicKey)
		if err != nil {
			m.handleAuthErrors(w, r.Context(), err)
			return
		}

		if !isValidAccessToken {
			EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: "Invalid access token"})
			return
		}

		userId, err := uuid.Parse(claims.Subject)
		if err != nil {
			EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
			m.Logger.Error(r.Context(), err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) LoginRequiredOrAllowAnonymous(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := m.GetAccessToken(r)
		if err != nil {
			EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
			m.Logger.Error(r.Context(), err.Error())
			return
		}

		// After getting the access token value, cover 4 scenarios
		// 1. Access token is empty and not allow anonymous
		// 2. Access token is empty and allow anonymous
		// 3. Access token is not empty and not allow anonymous
		// 4. Access token is not empty and allow anonymous
		// Scenarios 3 and 4 can be treated as a single scenario, require token to be valid

		// Scenario 1.
		if accessToken == "" && !m.Config.Server.AllowAnonymous {
			EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: "Login required"})
			return
		} else if accessToken == "" && m.Config.Server.AllowAnonymous { // Scenario 2
			ctx := context.WithValue(r.Context(), UserIdKey, uuid.Nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Scenario 3 and 4
		claims, isValidAccessToken, err := ValidateAccessToken(accessToken, &m.Config.Server.Auth.JwtEcdsaParsedKey.PublicKey)
		if err != nil {
			m.handleAuthErrors(w, r.Context(), err)
			return
		}

		if !isValidAccessToken {
			EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: "Invalid access token"})
			return
		}

		userId, err := uuid.Parse(claims.Subject)
		if err != nil {
			EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
			m.Logger.Error(r.Context(), err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) GetAccessToken(r *http.Request) (accessToken string, err error) {
	authHeaderValue := r.Header.Get(HeaderAuthorization)
	if authHeaderValue != "" {
		accessToken = strings.TrimPrefix(authHeaderValue, HeaderAuthorizationPrefix)
		return accessToken, nil
	}

	cookie, err := r.Cookie(CookieAccessTokenName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return "", err
	}

	if cookie != nil {
		accessToken = cookie.Value
		return accessToken, err
	}

	return "", nil
}

func (m *Middleware) handleAuthErrors(w http.ResponseWriter, ctx context.Context, err error) {
	if errors.Is(err, jwt.ErrECDSAVerification) ||
		errors.Is(err, jwt.ErrTokenMalformed) ||
		errors.Is(err, jwt.ErrTokenNotValidYet) ||
		errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
		errors.Is(err, jwt.ErrTokenUnverifiable) ||
		errors.Is(err, jwt.ErrTokenRequiredClaimMissing) {
		EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: "Invalid access token"})
		return
	}

	EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
	m.Logger.Error(ctx, err.Error())
}
