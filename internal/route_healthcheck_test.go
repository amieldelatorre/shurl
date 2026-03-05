package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/google/go-cmp/cmp"
)

type GetHealthCheckCase struct {
	Name               string
	TerminateDb        bool
	TerminateCache     bool
	ExpectedStatusCode int
	Expected           handlers.HealthCheckResponse
}

func TestHealthCheck(t *testing.T) {
	cacheOk := true
	cacheNOk := false
	cases := []GetHealthCheckCase{
		{
			Name:               "AllOk",
			TerminateDb:        false,
			TerminateCache:     false,
			ExpectedStatusCode: http.StatusOK,
			Expected: handlers.HealthCheckResponse{
				IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
				ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
				Database: handlers.DatabaseHealthCheck{
					Ok:      true,
					Version: DB_VERSION,
				},
				Cache: handlers.CacheHealthCheck{
					Enabled: true,
					Ok:      &cacheOk,
				},
			},
		},
		{
			Name:               "AllOkCacheDisabled",
			TerminateDb:        false,
			TerminateCache:     false,
			ExpectedStatusCode: http.StatusOK,
			Expected: handlers.HealthCheckResponse{
				IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
				ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
				Database: handlers.DatabaseHealthCheck{
					Ok:      true,
					Version: DB_VERSION,
				},
				Cache: handlers.CacheHealthCheck{
					Enabled: false,
					Ok:      nil,
				},
			},
		},
		{
			Name:               "DbNotOk",
			TerminateDb:        true,
			TerminateCache:     false,
			ExpectedStatusCode: http.StatusInternalServerError,
			Expected: handlers.HealthCheckResponse{
				IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
				ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
				Database: handlers.DatabaseHealthCheck{
					Ok:      false,
					Version: "0",
				},
				Cache: handlers.CacheHealthCheck{
					Enabled: true,
					Ok:      &cacheOk,
				},
				Errors: []string{"could not ping database", "could not get db version"},
			},
		},
		{
			Name:               "CacheNotOk",
			TerminateDb:        false,
			TerminateCache:     true,
			ExpectedStatusCode: http.StatusInternalServerError,
			Expected: handlers.HealthCheckResponse{
				IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
				ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
				Database: handlers.DatabaseHealthCheck{
					Ok:      true,
					Version: DB_VERSION,
				},
				Cache: handlers.CacheHealthCheck{
					Enabled: true,
					Ok:      &cacheNOk,
				},
				Errors: []string{"could not ping cache"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			deps := SetupDependencies(t, ctx, testCase.Expected.Cache.Enabled)

			defer func() {
				if err := deps.App.Server.Close(); err != nil {
					t.Fatal(err)
				}

				if !testCase.TerminateDb {
					if err := deps.Db.Container.Terminate(ctx); err != nil {
						t.Fatal(err)
					}
				}

				if testCase.Expected.Cache.Enabled && !testCase.TerminateCache {
					if err := deps.Cache.Container.Terminate(ctx); err != nil {
						t.Fatal(err)
					}
				}
			}()

			if testCase.TerminateDb {
				if err := deps.Db.Container.Terminate(ctx); err != nil {
					t.Fatal(err)
				}
			}

			if testCase.Expected.Cache.Enabled && testCase.TerminateCache {
				if err := deps.Cache.Container.Terminate(ctx); err != nil {
					t.Fatal(err)
				}
			}

			res, err := http.Get(deps.TestServer.URL + "/api/v1/health")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != testCase.ExpectedStatusCode {
				t.Errorf("expected status %d got %d", testCase.ExpectedStatusCode, res.StatusCode)
			}

			var health handlers.HealthCheckResponse
			if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
				t.Fatal("failed to decode body", err.Error())
			}
			err = res.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(testCase.Expected, health); diff != "" {
				t.Errorf("actual does not equal expected. diff: %s", diff)
			}
		})
	}
}

// func TestHealthCheckOk(t *testing.T) {
// 	ctx := context.Background()
// 	cacheok := true
// 	expected := handlers.HealthCheckResponse{
// 		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
// 		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
// 		Database: handlers.DatabaseHealthCheck{
// 			Ok:      true,
// 			Version: DB_VERSION,
// 		},
// 		Cache: handlers.CacheHealthCheck{
// 			Enabled: true,
// 			Ok:      &cacheok,
// 		},
// 	}

// 	deps := SetupDependencies(t, ctx, expected.Cache.Enabled)
// 	defer func() {
// 		if err := deps.App.Server.Close(); err != nil {
// 			t.Fatal(err)
// 		}

// 		if err := deps.Db.Container.Terminate(ctx); err != nil {
// 			t.Fatal(err)
// 		}
// 		if expected.Cache.Enabled {
// 			if err := deps.Cache.Container.Terminate(ctx); err != nil {
// 				t.Fatal(err)
// 			}
// 		}
// 	}()

// 	res, err := http.Get(deps.TestServer.URL + "/api/v1/health")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if res.StatusCode != http.StatusOK {
// 		t.Errorf("expected status %d got %d", http.StatusOK, res.StatusCode)
// 	}

// 	var health handlers.HealthCheckResponse
// 	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
// 		t.Fatal("failed to decode body", err.Error())
// 	}
// 	err = res.Body.Close()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := cmp.Diff(expected, health); diff != "" {
// 		t.Errorf("actual does not equal expected. diff: %s", diff)
// 	}
// }

// func TestHealthCheckOkNoCache(t *testing.T) {
// 	ctx := context.Background()
// 	expected := handlers.HealthCheckResponse{
// 		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
// 		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
// 		Database: handlers.DatabaseHealthCheck{
// 			Ok:      true,
// 			Version: DB_VERSION,
// 		},
// 		Cache: handlers.CacheHealthCheck{
// 			Enabled: false,
// 			Ok:      nil,
// 		},
// 	}

// 	deps := SetupDependencies(t, ctx, expected.Cache.Enabled)
// 	defer func() {
// 		if err := deps.App.Server.Close(); err != nil {
// 			t.Fatal(err)
// 		}

// 		if err := deps.Db.Container.Terminate(ctx); err != nil {
// 			t.Fatal(err)
// 		}
// 		if expected.Cache.Enabled {
// 			if err := deps.Cache.Container.Terminate(ctx); err != nil {
// 				t.Fatal(err)
// 			}
// 		}
// 	}()

// 	res, err := http.Get(deps.TestServer.URL + "/api/v1/health")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if res.StatusCode != http.StatusOK {
// 		t.Errorf("expected status %d got %d", http.StatusOK, res.StatusCode)
// 	}

// 	var health handlers.HealthCheckResponse
// 	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
// 		t.Fatal("failed to decode body", err.Error())
// 	}
// 	err = res.Body.Close()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := cmp.Diff(expected, health); diff != "" {
// 		t.Errorf("actual does not equal expected. diff: %s", diff)
// 	}
// }

// func TestHealthCheckDbNOk(t *testing.T) {
// 	ctx := context.Background()
// 	cacheok := true
// 	expected := handlers.HealthCheckResponse{
// 		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
// 		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
// 		Database: handlers.DatabaseHealthCheck{
// 			Ok:      false,
// 			Version: "0",
// 		},
// 		Cache: handlers.CacheHealthCheck{
// 			Enabled: true,
// 			Ok:      &cacheok,
// 		},
// 		Errors: []string{"could not ping database", "could not get db version"},
// 	}

// 	deps := SetupDependencies(t, ctx, expected.Cache.Enabled)
// 	defer func() {
// 		if err := deps.App.Server.Close(); err != nil {
// 			t.Fatal(err)
// 		}

// 		if expected.Cache.Enabled {
// 			if err := deps.Cache.Container.Terminate(ctx); err != nil {
// 				t.Fatal(err)
// 			}
// 		}
// 	}()

// 	if err := deps.Db.Container.Terminate(ctx); err != nil {
// 		t.Fatal(err)
// 	}

// 	res, err := http.Get(deps.TestServer.URL + "/api/v1/health")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if res.StatusCode != http.StatusInternalServerError {
// 		t.Errorf("expected status %d got %d", http.StatusInternalServerError, res.StatusCode)
// 	}

// 	var health handlers.HealthCheckResponse
// 	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
// 		t.Fatal("failed to decode body", err.Error())
// 	}
// 	err = res.Body.Close()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := cmp.Diff(expected, health); diff != "" {
// 		t.Errorf("actual does not equal expected. diff: %s", diff)
// 	}
// }

// func TestHealthCheckCacheNOk(t *testing.T) {
// 	ctx := context.Background()
// 	cacheok := false
// 	expected := handlers.HealthCheckResponse{
// 		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
// 		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
// 		Database: handlers.DatabaseHealthCheck{
// 			Ok:      true,
// 			Version: DB_VERSION,
// 		},
// 		Cache: handlers.CacheHealthCheck{
// 			Enabled: true,
// 			Ok:      &cacheok,
// 		},
// 		Errors: []string{"could not ping cache"},
// 	}

// 	deps := SetupDependencies(t, ctx, expected.Cache.Enabled)
// 	defer func() {
// 		if err := deps.App.Server.Close(); err != nil {
// 			t.Fatal(err)
// 		}

// 		if err := deps.Db.Container.Terminate(ctx); err != nil {
// 			t.Fatal(err)
// 		}
// 	}()

// 	if err := deps.Cache.Container.Terminate(ctx); err != nil {
// 		t.Fatal(err)
// 	}

// 	res, err := http.Get(deps.TestServer.URL + "/api/v1/health")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if res.StatusCode != http.StatusInternalServerError {
// 		t.Errorf("expected status %d got %d", http.StatusInternalServerError, res.StatusCode)
// 	}

// 	var health handlers.HealthCheckResponse
// 	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
// 		t.Fatal("failed to decode body", err.Error())
// 	}
// 	err = res.Body.Close()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := cmp.Diff(expected, health); diff != "" {
// 		t.Errorf("actual does not equal expected. diff: %s", diff)
// 	}
// }
