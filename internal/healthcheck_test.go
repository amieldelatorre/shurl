package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/test"
	"github.com/google/go-cmp/cmp"
)

func TestHealthCheckOk(t *testing.T) {
	ctx := context.Background()
	cacheok := true
	expected := handlers.HealthCheckResponse{
		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
		Database: handlers.DatabaseHealthCheck{
			Ok:      true,
			Version: test.DB_VERSION,
		},
		Cache: handlers.CacheHealthCheck{
			Enabled: true,
			Ok:      &cacheok,
		},
	}

	pg, err := test.GetPostgreSqlInstace(ctx)
	if err != nil {
		t.Fatal("could not start pg", err.Error())
	}
	defer func() {
		if err := pg.Container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	valkey, err := test.GetValkeyInstane(ctx)
	if err != nil {
		t.Fatal("could not start valkey", err.Error())
	}
	defer func() {
		if err := valkey.Container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	t.Setenv("SERVER_DOMAIN", test.SERVER_DOMAIN)
	t.Setenv("DATABASE_RUN_MIGRATIONS", "true")
	t.Setenv("DATABASE_DRIVER", test.DB_DRIVER)
	t.Setenv("DATABASE_HOST", pg.Host)
	t.Setenv("DATABASE_PORT", pg.Port)
	t.Setenv("DATABASE_NAME", test.DB_NAME)
	t.Setenv("DATABASE_USERNAME", test.DB_USERNAME)
	t.Setenv("DATABASE_PASSWORD", test.DB_PASSWORD)
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("CACHE_HOST", valkey.Host)
	t.Setenv("CACHE_PORT", valkey.Port)
	t.Setenv("SERVER_AUTH_JWT_KEY", test.JWT_KEY)

	app := NewApp("")
	ts := httptest.NewServer(app.Server.Handler)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d got %d", http.StatusOK, res.StatusCode)
	}

	var health handlers.HealthCheckResponse
	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatal("failed to decode body", err.Error())
	}
	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, health); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}
}

func TestHealthCheckOkNoCache(t *testing.T) {
	ctx := context.Background()
	expected := handlers.HealthCheckResponse{
		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
		Database: handlers.DatabaseHealthCheck{
			Ok:      true,
			Version: test.DB_VERSION,
		},
		Cache: handlers.CacheHealthCheck{
			Enabled: false,
			Ok:      nil,
		},
	}

	pg, err := test.GetPostgreSqlInstace(ctx)
	if err != nil {
		t.Fatal("could not start pg", err.Error())
	}
	defer func() {
		if err := pg.Container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	t.Setenv("SERVER_DOMAIN", test.SERVER_DOMAIN)
	t.Setenv("DATABASE_RUN_MIGRATIONS", "true")
	t.Setenv("DATABASE_DRIVER", test.DB_DRIVER)
	t.Setenv("DATABASE_HOST", pg.Host)
	t.Setenv("DATABASE_PORT", pg.Port)
	t.Setenv("DATABASE_NAME", test.DB_NAME)
	t.Setenv("DATABASE_USERNAME", test.DB_USERNAME)
	t.Setenv("DATABASE_PASSWORD", test.DB_PASSWORD)
	t.Setenv("CACHE_ENABLED", "false")
	t.Setenv("SERVER_AUTH_JWT_KEY", test.JWT_KEY)

	app := NewApp("")
	ts := httptest.NewServer(app.Server.Handler)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d got %d", http.StatusOK, res.StatusCode)
	}

	var health handlers.HealthCheckResponse
	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatal("failed to decode body", err.Error())
	}
	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, health); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}
}

func TestHealthCheckDbNOk(t *testing.T) {
	ctx := context.Background()
	cacheok := true
	expected := handlers.HealthCheckResponse{
		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
		Database: handlers.DatabaseHealthCheck{
			Ok:      false,
			Version: "0",
		},
		Cache: handlers.CacheHealthCheck{
			Enabled: true,
			Ok:      &cacheok,
		},
		Errors: []string{"could not ping database", "could not get db version"},
	}

	pg, err := test.GetPostgreSqlInstace(ctx)
	if err != nil {
		t.Fatal("could not start pg", err.Error())
	}

	valkey, err := test.GetValkeyInstane(ctx)
	if err != nil {
		t.Fatal("could not start valkey", err.Error())
	}
	defer func() {
		if err := valkey.Container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	t.Setenv("SERVER_DOMAIN", test.SERVER_DOMAIN)
	t.Setenv("DATABASE_RUN_MIGRATIONS", "true")
	t.Setenv("DATABASE_DRIVER", test.DB_DRIVER)
	t.Setenv("DATABASE_HOST", pg.Host)
	t.Setenv("DATABASE_PORT", pg.Port)
	t.Setenv("DATABASE_NAME", test.DB_NAME)
	t.Setenv("DATABASE_USERNAME", test.DB_USERNAME)
	t.Setenv("DATABASE_PASSWORD", test.DB_PASSWORD)
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("CACHE_HOST", valkey.Host)
	t.Setenv("CACHE_PORT", valkey.Port)
	t.Setenv("SERVER_AUTH_JWT_KEY", test.JWT_KEY)

	app := NewApp("")
	ts := httptest.NewServer(app.Server.Handler)
	defer ts.Close()
	if err := pg.Container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d got %d", http.StatusInternalServerError, res.StatusCode)
	}

	var health handlers.HealthCheckResponse
	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatal("failed to decode body", err.Error())
	}
	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, health); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}
}

func TestHealthCheckCacheNOk(t *testing.T) {
	ctx := context.Background()
	cacheok := false
	expected := handlers.HealthCheckResponse{
		IdempotencyKeyCleanupWorker: handlers.IdempotencyKeyCleanupWorkerHealthCheck{},
		ShortUrlCleanUpWorker:       handlers.ShortUrlCleanupWorkerHealthCheck{},
		Database: handlers.DatabaseHealthCheck{
			Ok:      true,
			Version: test.DB_VERSION,
		},
		Cache: handlers.CacheHealthCheck{
			Enabled: true,
			Ok:      &cacheok,
		},
		Errors: []string{"could not ping cache"},
	}

	pg, err := test.GetPostgreSqlInstace(ctx)
	if err != nil {
		t.Fatal("could not start pg", err.Error())
	}
	defer func() {
		if err := pg.Container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	valkey, err := test.GetValkeyInstane(ctx)
	if err != nil {
		t.Fatal("could not start valkey", err.Error())
	}

	t.Setenv("SERVER_DOMAIN", test.SERVER_DOMAIN)
	t.Setenv("DATABASE_RUN_MIGRATIONS", "true")
	t.Setenv("DATABASE_DRIVER", test.DB_DRIVER)
	t.Setenv("DATABASE_HOST", pg.Host)
	t.Setenv("DATABASE_PORT", pg.Port)
	t.Setenv("DATABASE_NAME", test.DB_NAME)
	t.Setenv("DATABASE_USERNAME", test.DB_USERNAME)
	t.Setenv("DATABASE_PASSWORD", test.DB_PASSWORD)
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("CACHE_HOST", valkey.Host)
	t.Setenv("CACHE_PORT", valkey.Port)
	t.Setenv("SERVER_AUTH_JWT_KEY", test.JWT_KEY)

	app := NewApp("")
	ts := httptest.NewServer(app.Server.Handler)
	defer ts.Close()
	if err := valkey.Container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d got %d", http.StatusInternalServerError, res.StatusCode)
	}

	var health handlers.HealthCheckResponse
	if err = json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatal("failed to decode body", err.Error())
	}
	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, health); diff != "" {
		t.Errorf("actual does not equal expected. diff: %s", diff)
	}
}
