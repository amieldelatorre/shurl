package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type LoginTestCase struct {
	Name               string
	Request            handlers.LoginRequest
	AllowLogin         bool
	SkipJsonHeader     bool
	ExpectedStatusCode int
	Expected           handlers.LoginResponse
}

func TestLogin(t *testing.T) {
	cases := []LoginTestCase{
		{
			Name: "LoginNotAllowed",
			Request: handlers.LoginRequest{
				Email:    "",
				Password: "",
			},
			SkipJsonHeader:     true,
			AllowLogin:         false,
			ExpectedStatusCode: http.StatusForbidden,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Login has been disabled by the administrator",
				},
			},
		},
		{
			Name: "NoJsonHeader",
			Request: handlers.LoginRequest{
				Email:    "",
				Password: "",
			},
			SkipJsonHeader:     true,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusBadRequest,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Endpoint requires header 'Content-Type' with value 'application/json'",
				},
			},
		},
		{
			Name: "Empty",
			Request: handlers.LoginRequest{
				Email:    "",
				Password: "",
			},
			SkipJsonHeader:     false,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusBadRequest,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag",
					"Key: 'LoginRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag",
				},
			},
		},
		{
			Name: "InvalidEmail",
			Request: handlers.LoginRequest{
				Email:    "adsfasdfadsf",
				Password: "123",
			},
			SkipJsonHeader:     false,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusBadRequest,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
				},
			},
		},
		{
			Name: "WrongPassword",
			Request: handlers.LoginRequest{
				Email:    "test1@example.invalid",
				Password: "123",
			},
			SkipJsonHeader:     false,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusUnauthorized,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Invalid credentials",
				},
			},
		},
		{
			Name: "EmailNotExist",
			Request: handlers.LoginRequest{
				Email:    "NotExist@example.invalid",
				Password: "123",
			},
			SkipJsonHeader:     false,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusUnauthorized,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Invalid credentials",
				},
			},
		},
		{
			Name: "SomeoneElsePassword",
			Request: handlers.LoginRequest{
				Email:    "test1@example.invalid",
				Password: "password123",
			},
			SkipJsonHeader:     false,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusUnauthorized,
			Expected: handlers.LoginResponse{
				Errors: []string{
					"Invalid credentials",
				},
			},
		},
		{
			Name: "HappyPath",
			Request: handlers.LoginRequest{
				Email:    "test1@example.invalid",
				Password: "password",
			},
			SkipJsonHeader:     false,
			AllowLogin:         true,
			ExpectedStatusCode: http.StatusCreated,
			Expected:           handlers.LoginResponse{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name+"WithCache", func(t *testing.T) {
			runLogin(t, tc, true)
		})
		t.Run(tc.Name+"NoCache", func(t *testing.T) {
			runLogin(t, tc, false)
		})
	}
}

func runLogin(t *testing.T, tc LoginTestCase, cacheEnabled bool) {
	ctx := context.Background()
	t.Setenv("SERVER_ALLOW_LOGIN", strconv.FormatBool(tc.AllowLogin))
	deps := SetupDependencies(t, ctx, cacheEnabled)

	rbody, err := json.Marshal(tc.Request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, deps.TestServer.URL+"/api/v1/auth/login", bytes.NewBuffer(rbody))
	if err != nil {
		t.Fatal(err)
	}

	if !tc.SkipJsonHeader {
		req.Header.Set(types.HeadersContentTypeKey, types.HeadersContentTypeJsonValue)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != tc.ExpectedStatusCode {
		t.Errorf("expected status %d got %d", tc.ExpectedStatusCode, res.StatusCode)
	}

	var loginResponse handlers.LoginResponse
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&loginResponse); err != nil {
		t.Error("failed to decode body", err.Error())
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(tc.Expected, loginResponse, cmpopts.IgnoreFields(handlers.LoginResponse{}, "AccessToken")); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}

	if tc.ExpectedStatusCode == http.StatusCreated {
		if loginResponse.AccessToken == nil {
			t.Errorf("access token is nil")
		}
	}
}
