package internal

import (
	"context"
	"net/http"
	"testing"
)

type RedirectionTestCase struct {
	Name               string
	slug               string
	ExpectedStatusCode int
	ExpectedHeaders    map[string]string
}

func TestRedirection(t *testing.T) {
	cases := []RedirectionTestCase{
		{
			Name:               "NotFound",
			slug:               "asdfadsasdfasdf",
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedHeaders:    map[string]string{},
		},
		{
			Name:               "Found",
			slug:               "tiLd",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
			ExpectedHeaders: map[string]string{
				"Location": "https://google.com",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name+"WithCache", func(t *testing.T) {
			runTestRedirect(t, tc, true)
		})
		t.Run(tc.Name+"NoCache", func(t *testing.T) {
			runTestRedirect(t, tc, false)
		})
	}
}

func runTestRedirect(t *testing.T, tc RedirectionTestCase, cacheEnabled bool) {
	ctx := context.Background()
	deps := SetupDependencies(t, ctx, cacheEnabled)
	req, err := http.NewRequest(http.MethodGet, deps.TestServer.URL+"/"+tc.slug, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != tc.ExpectedStatusCode {
		t.Errorf("expected status %d got %d", tc.ExpectedStatusCode, res.StatusCode)
	}

	for k, v := range tc.ExpectedHeaders {
		actualHeader := res.Header.Get(k)
		if actualHeader != v {
			t.Errorf("expected %s header value %s, got %s", k, v, actualHeader)
		}
	}
}
