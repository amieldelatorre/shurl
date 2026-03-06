package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	Expected              types.ShortUrlResponse
}

var (
	usedIdempotencyKey = uuid.MustParse("019cc05a-72a5-7479-a1dd-0105df4fc6c4")
	validUserUuid      = uuid.MustParse("019cc05a-7415-7528-8c5a-e0487fad449c")
)

func TestPostShortUrl(t *testing.T) {
	happyPathUrl := "https://google.com"
	var ttlLessThanMin uint32 = 899
	var ttlOnMin uint32 = 900
	ttlOnAnonymousMax := handlers.MaxAnonymousShortUrlTtl
	var ttlGreaterThanAnonymousMax uint32 = 604801
	ttlOnAuthenticatedMax := handlers.MaxAuthenticatedShortUrlTtl
	var ttlGreaterThanAuthenticatedMax uint32 = 2629747

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
			Expected: types.ShortUrlResponse{
				Errors: []string{"Login required"},
			},
		},
		{
			Name: "AnonymousLessThanMin",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlLessThanMin,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.ShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.TTL' Error:Field validation for 'TTL' failed on the 'min' tag"},
			},
		},
		{
			Name: "AnonymousGreaterThanMax",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlGreaterThanAnonymousMax,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.ShortUrlResponse{
				Errors: []string{"anonymous short urls can only be up to 604800 seconds"},
			},
		},
		{
			Name: "AnonymousOnMax",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlOnAnonymousMax,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.ShortUrlResponse{
				DestinationUrl: &happyPathUrl,
			},
		},
		{
			Name: "AnonymousOnMin",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlOnMin,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           nil,
			UseCookie:             false,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.ShortUrlResponse{
				DestinationUrl: &happyPathUrl,
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.DestinationUrl' Error:Field validation for 'DestinationUrl' failed on the 'url' tag"},
			},
		},
		{
			Name: "AuthenticatedLessThanMin",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlLessThanMin,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.ShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.TTL' Error:Field validation for 'TTL' failed on the 'min' tag"},
			},
		},
		{
			Name: "AuthenticatedGreaterThanMax",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlGreaterThanAuthenticatedMax,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: types.ShortUrlResponse{
				Errors: []string{"Key: 'PostShortUrlRequest.TTL' Error:Field validation for 'TTL' failed on the 'max' tag"},
			},
		},
		{
			Name: "AuthenticatedOnMax",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlOnAuthenticatedMax,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.ShortUrlResponse{
				DestinationUrl: &happyPathUrl,
				UserId:         &validUserUuid,
			},
		},
		{
			Name: "AuthenticatedOnMin",
			Request: handlers.PostShortUrlRequest{
				DestinationUrl: "https://google.com",
				TTL:            &ttlOnMin,
			},
			AllowAnonymous:        true,
			SkipIdempotencyKey:    false,
			SkipJsonHeader:        false,
			UseIdempotencyKeyUuid: nil,
			UseUserUuid:           &validUserUuid,
			UseCookie:             true,
			UseHeader:             false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: types.ShortUrlResponse{
				DestinationUrl: &happyPathUrl,
				UserId:         &validUserUuid,
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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
			Expected: types.ShortUrlResponse{
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

	var shortUrlPostResponse types.ShortUrlResponse
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&shortUrlPostResponse); err != nil {
		t.Error("failed to decode body", err.Error())
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(tc.Expected, shortUrlPostResponse, cmpopts.IgnoreFields(types.ShortUrlResponse{}, "CreatedAt", "ExpiresAt", "Slug", "Id", "Url")); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}

	if tc.ExpectedStatusCode == http.StatusCreated && tc.UseUserUuid != nil {
		if shortUrlPostResponse.UserId == nil {
			t.Errorf("user id is nil")
		}
	}
}

type GetShortUrlsByUserIdCase struct {
	Name               string
	Page               int
	SkipPage           bool
	Size               int
	SkipSize           bool
	UserUuid           uuid.UUID
	SkipAccessToken    bool
	ExpectedStatusCode int
	Expected           handlers.GetShortUrlsByUserIdResponse
}

func TestGetShortUrlsByUserId(t *testing.T) {
	t.Parallel()

	expect1Id := uuid.MustParse("019cc05b-d0e6-764d-a207-60cb9fd4d147")
	expect1Destinationurl := "https://google.com"
	expect1Slug := "zzM0ofu"
	expect1Url := "http://localhost:8080/zzM0ofu"

	expect2Id := uuid.MustParse("019cc05b-c45d-76f9-ab03-02af299e76ea")
	expect2Destinationurl := "https://google.com"
	expect2Slug := "4kJe27"
	expect2Url := "http://localhost:8080/4kJe27"

	expect3Id := uuid.MustParse("019cc05b-b0ca-7bf2-863f-2356491c227d")
	expect3Destinationurl := "https://google.com"
	expect3Slug := "S0VieOF"
	expect3Url := "http://localhost:8080/S0VieOF"

	cases := []GetShortUrlsByUserIdCase{
		{
			Name:               "LoginRequired",
			Page:               -1,
			Size:               0,
			UserUuid:           validUserUuid,
			SkipAccessToken:    true,
			ExpectedStatusCode: http.StatusUnauthorized,
			Expected: handlers.GetShortUrlsByUserIdResponse{
				Errors: []string{"Login required"},
			},
		},
		{
			Name:               "InvalidPage",
			Page:               -1,
			Size:               0,
			UserUuid:           validUserUuid,
			ExpectedStatusCode: http.StatusBadRequest,
			Expected: handlers.GetShortUrlsByUserIdResponse{
				Errors: []string{handlers.PageQueryParamError},
			},
		},
		{
			Name:               "InvalidSize",
			Page:               0,
			Size:               0,
			UserUuid:           validUserUuid,
			ExpectedStatusCode: http.StatusBadRequest,
			Expected: handlers.GetShortUrlsByUserIdResponse{
				Errors: []string{handlers.SizeQueryParamError},
			},
		},
		{
			Name:               "Expect1",
			Page:               0,
			Size:               1,
			UserUuid:           validUserUuid,
			ExpectedStatusCode: http.StatusOK,
			Expected: handlers.GetShortUrlsByUserIdResponse{
				Items: []types.ShortUrlResponse{
					{
						Id:             &expect1Id,
						DestinationUrl: &expect1Destinationurl,
						Slug:           &expect1Slug,
						Url:            expect1Url,
						UserId:         &validUserUuid,
					},
				},
			},
		},
		{
			Name:               "ExpectNoneOffsetOvershoot",
			Page:               3,
			Size:               20,
			UserUuid:           validUserUuid,
			ExpectedStatusCode: http.StatusOK,
			Expected: handlers.GetShortUrlsByUserIdResponse{
				Items: []types.ShortUrlResponse{},
			},
		},
		{
			Name:               "Expect3Normal",
			Page:               -1,
			SkipPage:           true,
			Size:               -1,
			SkipSize:           true,
			UserUuid:           validUserUuid,
			ExpectedStatusCode: http.StatusOK,
			Expected: handlers.GetShortUrlsByUserIdResponse{
				Items: []types.ShortUrlResponse{
					{
						Id:             &expect1Id,
						DestinationUrl: &expect1Destinationurl,
						Slug:           &expect1Slug,
						Url:            expect1Url,
						UserId:         &validUserUuid,
					},
					{
						Id:             &expect2Id,
						DestinationUrl: &expect2Destinationurl,
						Slug:           &expect2Slug,
						Url:            expect2Url,
						UserId:         &validUserUuid,
					},
					{
						Id:             &expect3Id,
						DestinationUrl: &expect3Destinationurl,
						Slug:           &expect3Slug,
						Url:            expect3Url,
						UserId:         &validUserUuid,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name+"WithCache", func(t *testing.T) {
			t.Parallel()
			runTestGetShortUrlsById(t, tc, true)
		})
		t.Run(tc.Name+"NoCache", func(t *testing.T) {
			t.Parallel()
			runTestGetShortUrlsById(t, tc, false)
		})
	}
}

func runTestGetShortUrlsById(t *testing.T, tc GetShortUrlsByUserIdCase, cacheEnabled bool) {
	ctx := context.Background()
	deps := SetupDependencies(t, ctx, cacheEnabled)

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

	req, err := http.NewRequest(http.MethodGet, deps.TestServer.URL+"/api/v1/me/shorturl", nil)
	if err != nil {
		t.Fatal(err)
	}
	queryValues := req.URL.Query()
	if !tc.SkipPage {
		queryValues.Add("page", strconv.Itoa(tc.Page))
	}
	if !tc.SkipSize {
		queryValues.Add("size", strconv.Itoa(tc.Size))
	}
	req.URL.RawQuery = queryValues.Encode()

	client := &http.Client{}

	accessToken := CreateAccessToken(t, deps.App.Config.Server.Auth, 12, &tc.UserUuid, true)
	if !tc.SkipAccessToken {
		req.Header.Add(handlers.HeaderAuthorization, fmt.Sprintf("Bearer %s", accessToken))
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != tc.ExpectedStatusCode {
		t.Errorf("expected status %d got %d", tc.ExpectedStatusCode, res.StatusCode)
	}

	var response handlers.GetShortUrlsByUserIdResponse
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&response); err != nil {
		t.Error("failed to decode body", err.Error())
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(tc.Expected, response, cmpopts.IgnoreFields(types.ShortUrlResponse{}, "CreatedAt", "ExpiresAt")); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}
}

type DeleteShortUrlByIdCase struct {
	Name               string
	ShortUrlIdToDelete string
	UserId             uuid.UUID
	SkipAccessToken    bool
	ExpectedStatusCode int
	ExpectedErrors     types.ErrorResponse
}

func TestDeleteShortUrlById(t *testing.T) {
	t.Parallel()

	validShortUrlId := "019cc05b-d0e6-764d-a207-60cb9fd4d147"
	otherUserShortUrlId := "019cbb9b-b28c-7c35-9dc0-8f3c553ca432"
	nullUserShortUrlId := "019cc1c7-d1f0-734f-a2b7-a5ee16fbad0b"

	cases := []DeleteShortUrlByIdCase{
		{
			Name:               "HappyPath",
			ShortUrlIdToDelete: validShortUrlId,
			UserId:             validUserUuid,
			SkipAccessToken:    false,
			ExpectedStatusCode: http.StatusNoContent,
			ExpectedErrors:     types.ErrorResponse{},
		},
		{
			Name:               "NotLoggedIn",
			ShortUrlIdToDelete: validShortUrlId,
			UserId:             validUserUuid,
			SkipAccessToken:    true,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:               "DeleteOtherUserShortUrl",
			ShortUrlIdToDelete: otherUserShortUrlId,
			UserId:             validUserUuid,
			SkipAccessToken:    false,
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Name:               "DeleteAnonymousShortUrl",
			ShortUrlIdToDelete: nullUserShortUrlId,
			UserId:             validUserUuid,
			SkipAccessToken:    false,
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Name:               "DeleteNotExistentShortUrl",
			ShortUrlIdToDelete: uuid.Nil.String(),
			UserId:             validUserUuid,
			SkipAccessToken:    false,
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Name:               "InvalidUuid",
			ShortUrlIdToDelete: "sd",
			UserId:             validUserUuid,
			SkipAccessToken:    false,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedErrors: types.ErrorResponse{
				Errors: []string{"Short url id provided is not a valid uuid"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name+"WithCache", func(t *testing.T) {
			t.Parallel()
			runDeleteShortUrlById(t, tc, true)
		})
		t.Run(tc.Name+"NoCache", func(t *testing.T) {
			t.Parallel()
			runDeleteShortUrlById(t, tc, false)
		})
	}

}

func runDeleteShortUrlById(t *testing.T, tc DeleteShortUrlByIdCase, cacheEnabled bool) {
	ctx := context.Background()
	deps := SetupDependencies(t, ctx, cacheEnabled)
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

	req, err := http.NewRequest(http.MethodDelete, deps.TestServer.URL+"/api/v1/me/shorturl/"+tc.ShortUrlIdToDelete, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	accessToken := CreateAccessToken(t, deps.App.Config.Server.Auth, 12, &tc.UserId, true)
	if !tc.SkipAccessToken {
		req.Header.Add(handlers.HeaderAuthorization, fmt.Sprintf("Bearer %s", accessToken))
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != tc.ExpectedStatusCode {
		t.Errorf("expected status %d got %d", tc.ExpectedStatusCode, res.StatusCode)
	}

	if len(tc.ExpectedErrors.Errors) > 0 {
		var response types.ErrorResponse
		decoder := json.NewDecoder(res.Body)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&response); err != nil {
			t.Error("failed to decode body", err.Error())
		}

		err = res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(tc.ExpectedErrors, response); diff != "" {
			t.Errorf("actual does not equal expected. diff: %s", diff)
		}
	}

}
