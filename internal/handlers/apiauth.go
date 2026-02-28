package handlers

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

const (
	invalidCredentialsMessage        = "Invalid credentials"
	dummyPassword                    = "DUMMY_CREDENTIALS_FOR_CONSTANT_TIME_COMPARE"
	jwtTokenValidHours               = 24
	HeaderXAuthMethodWanted   string = "X-Auth-Method-Wanted"
)

type AuthMethod string

const (
	AuthMethodCookie AuthMethod = "cookie"
	AuthMethodJson   AuthMethod = "json"
)

type ApiAuthHandler struct {
	Logger            utils.CustomJsonLogger
	Config            config.Config
	Db                db.DbContext
	dummyPasswordHash string
}

func NewApiAuthHandler(logger utils.CustomJsonLogger, config config.Config, dbContext db.DbContext) (ApiAuthHandler, error) {
	dummyPasswordHash, err := argon2id.CreateHash(dummyPassword, argon2idParams)
	if err != nil {
		return ApiAuthHandler{}, err
	}

	return ApiAuthHandler{Logger: logger, Config: config, Db: dbContext, dummyPasswordHash: dummyPasswordHash}, nil
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type JwtClaims struct {
	jwt.RegisteredClaims
}

func (h *ApiAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorCode, message := parseJsonDecodeError(err)
		EncodeResponse[types.ErrorResponse](w, errorCode, types.ErrorResponse{Error: message})
		if errorCode == http.StatusInternalServerError {
			h.Logger.Error(r.Context(), "Server error when parsing json body. error: %v", "error", err.Error())
		}
		return
	}

	validate := validator.New()
	err = validate.Struct(req)
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

	// Get user
	user, err := h.Db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	if user == nil {
		_, _ = argon2id.ComparePasswordAndHash("HERE_FOR_PREVENTING_USER_ENUMERATION", h.dummyPasswordHash)
		// disregard error response because we're going to send an unauthorized anyway

		EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: invalidCredentialsMessage})
		// TODO: Add IP address
		h.Logger.Warn(r.Context(), "Failed login attempt")
		return
	}

	passwordMatch, err := argon2id.ComparePasswordAndHash(req.Password, user.PasswordHash)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	if !passwordMatch {
		EncodeResponse[types.ErrorResponse](w, http.StatusUnauthorized, types.ErrorResponse{Error: invalidCredentialsMessage})
		// TODO: Add IP address
		h.Logger.Warn(r.Context(), "failed login attempt")
		return
	}

	now := time.Now()
	expiresAt := now.Add(jwtTokenValidHours * time.Hour)
	claims := JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Id.String(),
			Issuer:    h.Config.Server.Auth.JwtIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	signedToken, err := token.SignedString(h.Config.Server.Auth.JwtEcdsaParsedKey)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}

	authMethodWanted := r.Header.Get(HeaderXAuthMethodWanted)
	switch authMethodWanted {
	case string(AuthMethodCookie):
		cookie := &http.Cookie{
			Name:     CookieAccessTokenName,
			Value:    signedToken,
			Path:     "/",
			MaxAge:   jwtTokenValidHours * 60 * 60, // x * minutes in an hour * seconds in a minute
			HttpOnly: true,
			Secure:   h.Config.Server.HttpsEnabled,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusCreated)
	case string(AuthMethodJson):
		EncodeResponse[loginResponse](w, http.StatusCreated, loginResponse{AccessToken: signedToken})
	default:
		EncodeResponse[loginResponse](w, http.StatusCreated, loginResponse{AccessToken: signedToken})
	}

	// TODO: Add IP address
	h.Logger.Warn(r.Context(), "login successful")
}

type ValidateResponse struct {
	Ok bool `json:"ok"`
}

func (h *ApiAuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	// if they've made it this far, its ok
	EncodeResponse[ValidateResponse](w, http.StatusOK, ValidateResponse{Ok: true})
}

func ValidateAccessToken(token string, publicKey *ecdsa.PublicKey) (*JwtClaims, bool, error) {
	claims := &JwtClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		if t.Method.Alg() != jwt.SigningMethodES512.Name {
			return nil, errors.New("unexpected signing method")
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, false, err
	}

	if !parsedToken.Valid {
		return nil, false, nil
	}

	return claims, true, nil
}
