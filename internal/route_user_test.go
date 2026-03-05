package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

type PostUserTestCase struct {
	Name                  string
	Request               handlers.PostUserRequest
	SkipIdempotencyHeader bool
	SkipJsonHeader        bool
	ExpectedStatusCode    int
	Expected              handlers.PostUserResponse
}

func TestPostUserTestCases(t *testing.T) {
	happyPathId, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	happyPathUsername := "new1"
	happyPathEmail := "new1@example.invalid"
	now := time.Now()

	cases := []PostUserTestCase{
		{
			Name: "MissingIdempotencyKey",
			Request: handlers.PostUserRequest{
				Username:        "",
				Email:           "",
				Password:        "",
				ConfirmPassword: "",
			},
			SkipIdempotencyHeader: true,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Missing uuidv7 idempotency key header 'X-Idempotency-Key'",
				},
			},
		},
		{
			Name: "MissingJsonHeader",
			Request: handlers.PostUserRequest{
				Username:        "",
				Email:           "",
				Password:        "",
				ConfirmPassword: "",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        true,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Endpoint required header 'Content-Type' with value 'application/json'",
				},
			},
		},
		{
			Name: "ValidationErrorsAllEmpty",
			Request: handlers.PostUserRequest{
				Username:        "",
				Email:           "",
				Password:        "",
				ConfirmPassword: "",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag",
					"Key: 'PostUserRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag",
					"Key: 'PostUserRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag",
					"Key: 'PostUserRequest.ConfirmPassword' Error:Field validation for 'ConfirmPassword' failed on the 'required' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsAllInvalid",
			Request: handlers.PostUserRequest{
				Username:        "ne",
				Email:           "aaaa",
				Password:        "1",
				ConfirmPassword: "2",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.Username' Error:Field validation for 'Username' failed on the 'min' tag",
					"Key: 'PostUserRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
					"Key: 'PostUserRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag",
					"Key: 'PostUserRequest.ConfirmPassword' Error:Field validation for 'ConfirmPassword' failed on the 'eqfield' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsPasswordNotLongEnough",
			Request: handlers.PostUserRequest{
				Username:        "new1",
				Email:           "new1@example.invalid",
				Password:        "passwor",
				ConfirmPassword: "passwor",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsPasswordsNotMatch",
			Request: handlers.PostUserRequest{
				Username:        "new1",
				Email:           "new1@example.invalid",
				Password:        "password",
				ConfirmPassword: "passwor",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.ConfirmPassword' Error:Field validation for 'ConfirmPassword' failed on the 'eqfield' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsUsernameExists",
			Request: handlers.PostUserRequest{
				Username:        "test1",
				Email:           "new1@example.invalid",
				Password:        "password",
				ConfirmPassword: "password",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Username or email already exists",
				},
			},
		},
		{
			Name: "ValidationErrorsUsernameInvalidCharSpace",
			Request: handlers.PostUserRequest{
				Username:        "tes            t1",
				Email:           "new1@example.invalid",
				Password:        "password",
				ConfirmPassword: "password",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.Username' Error:Field validation for 'Username' failed on the 'alphanum' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsUsernameInvalidChar",
			Request: handlers.PostUserRequest{
				Username:        "test1&",
				Email:           "new1@example.invalid",
				Password:        "password",
				ConfirmPassword: "password",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.Username' Error:Field validation for 'Username' failed on the 'alphanum' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsUsernameInvalidUppercase",
			Request: handlers.PostUserRequest{
				Username:        "UPPER",
				Email:           "new1@example.invalid",
				Password:        "password",
				ConfirmPassword: "password",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Key: 'PostUserRequest.Username' Error:Field validation for 'Username' failed on the 'lowercase' tag",
				},
			},
		},
		{
			Name: "ValidationErrorsEmailExists",
			Request: handlers.PostUserRequest{
				Username:        "new1",
				Email:           "test1@example.invalid",
				Password:        "password",
				ConfirmPassword: "password",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusBadRequest,
			Expected: handlers.PostUserResponse{
				Errors: []string{
					"Username or email already exists",
				},
			},
		},
		{
			Name: "HappyPath",
			Request: handlers.PostUserRequest{
				Username:        "new1",
				Email:           "new1@example.invalid",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			SkipIdempotencyHeader: false,
			SkipJsonHeader:        false,
			ExpectedStatusCode:    http.StatusCreated,
			Expected: handlers.PostUserResponse{
				Id:        &happyPathId,
				Username:  &happyPathUsername,
				Email:     &happyPathEmail,
				CreatedAt: &now,
				UpdatedAt: &now,
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name+"WithCache", func(t *testing.T) {
			runPostUserTest(t, testCase, true)
		})
		t.Run(testCase.Name+"NoCache", func(t *testing.T) {
			runPostUserTest(t, testCase, false)
		})
	}
}

// for tests to be automatically picked up, file needs to be named *_test.go and functions need to be named in the pattern TestXxx
func runPostUserTest(t *testing.T, testCase PostUserTestCase, cacheEnabled bool) {
	ctx := context.Background()
	t.Setenv("SERVER_ALLOW_REGISTRATION", "true")
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

	rbody, err := json.Marshal(testCase.Request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, deps.TestServer.URL+"/api/v1/user", bytes.NewBuffer(rbody))
	if err != nil {
		t.Fatal(err)
	}

	if !testCase.SkipIdempotencyHeader {
		iKey, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set(types.HeadersIdempotencyKey, iKey.String())
	}

	if !testCase.SkipJsonHeader {
		req.Header.Set(types.HeadersContentTypeKey, types.HeadersContentTypeJsonValue)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != testCase.ExpectedStatusCode {
		t.Errorf("expected status %d got %d", testCase.ExpectedStatusCode, res.StatusCode)
	}

	var postUserResponse handlers.PostUserResponse
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&postUserResponse); err != nil {
		t.Error("failed to decode body", err.Error())
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(testCase.Expected, postUserResponse, cmpopts.IgnoreFields(handlers.PostUserResponse{}, "Id", "CreatedAt", "UpdatedAt")); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}

	if testCase.ExpectedStatusCode == http.StatusCreated {
		if postUserResponse.Id == nil {
			t.Errorf("id is nil")
		}
		if postUserResponse.CreatedAt == nil {
			t.Errorf("created at is nil")
		}
		if postUserResponse.UpdatedAt == nil {
			t.Errorf("updated at is nil")
		}
	}
}
