package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

type PostShortUrlTestCase struct {
	Name                  string
	Request               handlers.PostShortUrlRequest
	AllowAnonymous        bool
	SkipIdempotencyKey    bool
	SkipJsonHeader        bool
	UseIdempotencyKeyUuid *uuid.UUID
	UseUserUuid           *uuid.UUID
	UseCookie             bool
	UseHeader             bool
	UseInvalidAccessToken bool
	UseExpiredAccessToken bool
	ExpectedStatusCode    int
	Expected              types.CreateShortUrlResponse
}

func TestPostShortUrl(t *testing.T) {
	usedIdempotencyKey := uuid.MustParse("019cc05a-72a5-7479-a1dd-0105df4fc6c4")
	validUserUuid := uuid.MustParse("019cc05a-7415-7528-8c5a-e0487fad449c")
	happyPathUrl := "https://google.com"

	t.Parallel()
	cases := []PostShortUrlTestCase{
		{
			Name: "AnonymousNotAllowed",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    true,
			SkipJsonHeader:        true,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Login required"},
			},
		},
		{
			Name: "AnonymousAllowedMissingIdempotencyKey",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "notvalid",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    true,
			SkipJsonHeader:        true,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Endpoint requires header 'Content-Type' with value 'application/json'"},
			},
		},
		{
			Name: "AnonymousAllowedMissingIdempotencyKey",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "notvalid",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    true,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Missing uuidv7 idempotency key header 'X-Idempotency-Key'"},
			},
		},
		{
			Name: "AnonymousAllowedIdempotencyKeyAlreadyUsed",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://gmail.com",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: &usedIdempotencyKey,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"X-Idempotency-Key header value has already been used"},
			},
		},
		{
			Name: "AnonymousAllowedEmptyUrl",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.DestinationUrl' Error:Field validation for 'DestinationUrl' failed on the 'required' tag"},
			},
		},
		{
			Name: "AnonymousAllowedInvalidUrl",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.DestinationUrl' Error:Field validation for 'DestinationUrl' failed on the 'url' tag"},
			},
		},
		{
			Name: "InvalidAccessTokenFromHeader",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             false,
			UseHeader:             true,
			UseInvalidAccessToken: true,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Invalid access token"},
			},
		},
		{
			Name: "InvalidAccessTokenFromCookie",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			UseInvalidAccessToken: true,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Invalid access token"},
			},
		},
		{
			Name: "InvalidAccessTokenFromHeaderAndCookie",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			UseInvalidAccessToken: true,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Invalid access token"},
			},
		},
		{
			Name: "ExpiredAccessTokenFromHeader",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             false,
			UseHeader:             true,
			UseExpiredAccessToken: true,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Invalid access token"},
			},
		},
		{
			Name: "ExpiredAccessTokenFromCookie",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			UseExpiredAccessToken: true,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Invalid access token"},
			},
		},
		{
			Name: "ExpiredAccessTokenFromHeaderAndCookie",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			UseExpiredAccessToken: true,
			ExpectedStatusCode:    http.StatusUnauthorized,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Invalid access token"},
			},
		},
		{
			Name: "MissingIdempotencyKey",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "notvalid",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    true,
			SkipJsonHeader:        true,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Endpoint requires header 'Content-Type' with value 'application/json'"},
			},
		},
		{
			Name: "MissingIdempotencyKey",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "notvalid",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    true,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Missing uuidv7 idempotency key header 'X-Idempotency-Key'"},
			},
		},
		{
			Name: "IdempotencyKeyAlreadyUsed",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://gmail.com",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: &usedIdempotencyKey,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"X-Idempotency-Key header value has already been used"},
			},
		},
		{
			Name: "EmptyUrl",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.DestinationUrl' Error:Field validation for 'DestinationUrl' failed on the 'required' tag"},
			},
		},
		{
			Name: "InvalidUrl",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "e",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.CreateShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.DestinationUrl' Error:Field validation for 'DestinationUrl' failed on the 'url' tag"},
			},
		},
		{
			Name: "HappyPathAuthenticated",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
			},
			AllowAnonymous:        false,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.CreateShortUrlResponse{
				DestinationUrl: &happyPathUrl,
				UserId:         &validUserUuid,
			},
		},
		{
			Name: "HappyPathAnonymous",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.CreateShortUrlResponse{
				DestinationUrl: &happyPathUrl,
			},
		},
		{
			Name: "HappyPathAuthenticatedAllowAnonymous",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.CreateShortUrlResponse{
				DestinationUrl: &happyPathUrl,
				UserId:         &validUserUuid,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name+"WithCache", func(t *testing.T) {
			t.Parallel()
			runTestPostShortUrl(t, tc, true)
		})
		t.Run(tc.Name+"NoCache", func(t *testing.T) {
			t.Parallel()
			runTestPostShortUrl(t, tc, false)
		})
	}
}

func runTestPostShortUrl(t *testing.T, tc PostShortUrlTestCase, cacheEnabled bool) {
	ctx := context.Background()
	deps := SetupDependencies(t, ctx, cacheEnabled)
	deps.App.Config.Server.AllowAnonymous = tc.AllowAnonymous

	defer func() {
		if err := deps.App.Server.Close(); err != nil {
			t.Fatal(err)
		}

		if err := deps.Db.Container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}

		if cacheEnabled {
			if err := deps.Cache.Container.Terminate(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}()

	var accessToken string
	if tc.UseUserUuid != nil {
		var h int
		if tc.UseExpiredAccessToken {
			h = -12
		} else {
			h = 12
		}
		accessToken = CreateAccessToken(t, deps.App.Config.Server.Auth, h, tc.UseUserUuid, !tc.UseInvalidAccessToken)
	}

	rbody, err := json.Marshal(tc.Request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, deps.TestServer.URL+"/api/v1/shorturl", bytes.NewBuffer(rbody))
	if err != nil {
		t.Fatal(err)
	}

	if !tc.SkipJsonHeader {
		req.Header.Set(types.HeadersContentTypeKey, types.HeadersContentTypeJsonValue)
	}

	if tc.UseHeader && tc.UseUserUuid != nil {
		req.Header.Add(handlers.HeaderAuthorization, fmt.Sprintf("Bearer %s", accessToken))
	}

	if !tc.SkipIdempotencyKey {
		var key uuid.UUID
		if tc.UseIdempotencyKeyUuid != nil {
			key = *tc.UseIdempotencyKeyUuid
		} else {
			key, err = uuid.NewV7()
			if err != nil {
				t.Fatal(err)
			}
		}
		req.Header.Add(types.HeadersIdempotencyKey, key.String())
	}

	if tc.UseCookie && tc.UseUserUuid != nil {
		cookie := &http.Cookie{
			Name:     handlers.CookieAccessTokenName,
			Value:    accessToken,
			Path:     "/",
			MaxAge:   12 * 60 * 60, // x * minutes in an hour * seconds in a minute
			Expires:  time.Now().Add(12 * time.Hour),
			HttpOnly: true,
			Secure:   deps.App.Config.Server.HttpsEnabled,
			SameSite: http.SameSiteStrictMode,
		}

		req.AddCookie(cookie)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != tc.ExpectedStatusCode {
		t.Errorf("expected status %d got %d", tc.ExpectedStatusCode, res.StatusCode)
	}

	var shortUrlPostResponse types.CreateShortUrlResponse
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&shortUrlPostResponse); err != nil {
		t.Error("failed to decode body", err.Error())
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(tc.Expected, shortUrlPostResponse, cmpopts.IgnoreFields(types.CreateShortUrlResponse{}, "CreatedAt", "ExpiresAt", "Slug", "Id", "Url")); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}

	if tc.ExpectedStatusCode == http.StatusCreated && tc.UseUserUuid != nil {
		if shortUrlPostResponse.UserId == nil {
			t.Errorf("user id is nil")
		}
	}
}
