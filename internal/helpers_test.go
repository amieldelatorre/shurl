package internal

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
)

type Db struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

func (d *Db) Init(ctx context.Context) error {
	status, out, err := d.Container.Exec(ctx, []string{"psql", "-U", DB_USERNAME, "-d", DB_NAME, "-f", TEST_DATA_PATH})
	if err != nil {
		return err
	}
	if status != 0 {
		outb, err := io.ReadAll(out)
		if err != nil {
			return err
		}

		return errors.New(string(outb))
	}

	return nil
}

type Cache struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

type Dependencies struct {
	Db         Db
	Cache      Cache
	App        App
	TestServer *httptest.Server
}

const (
	DB_VERSION     = "20260304055040"
	DB_NAME        = "shurl"
	DB_USERNAME    = "shurl"
	DB_PASSWORD    = "password"
	TEST_DATA_PATH = "/tmp/testdata.sql"
)

func GetValkeyInstane(ctx context.Context) (Cache, error) {
	valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:9.0.3")
	if err != nil {
		return Cache{}, err
	}

	host, err := valkeyContainer.Host(ctx)
	if err != nil {
		return Cache{}, err
	}

	port, err := valkeyContainer.MappedPort(ctx, "6379")
	if err != nil {
		return Cache{}, err
	}

	return Cache{Container: valkeyContainer, Host: host, Port: port.Port()}, nil
}

func GetPostgreSqlInstace(ctx context.Context) (Db, error) {
	// testData, err := os.ReadFile("test/testdata.sql")
	// if err != nil {
	// 	return Db{}, err
	// }

	postgresContainer, err := tcpostgres.Run(ctx,
		"postgres:18.3",
		tcpostgres.BasicWaitStrategies(),
		tcpostgres.WithDatabase(DB_NAME),
		tcpostgres.WithUsername(DB_USERNAME),
		tcpostgres.WithPassword(DB_PASSWORD),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      "test/testdata.sql",
			ContainerFilePath: TEST_DATA_PATH,
			FileMode:          0o644,
		},
		),
	)
	if err != nil {
		return Db{}, err
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return Db{}, err
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return Db{}, err
	}

	return Db{Container: postgresContainer, Host: host, Port: port.Port()}, nil
}

func SetupDependencies(t *testing.T, ctx context.Context, enableCache bool) Dependencies {
	db, err := GetPostgreSqlInstace(ctx)
	if err != nil {
		t.Fatal(err)
	}

	tempLogger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)

	config, err := config.LoadConfig("test/baseconf.yaml")
	if err != nil {
		tempLogger.ErrorExit(ctx, err.Error())
	}
	config.Database.Host = db.Host
	config.Database.Port = db.Port

	var cache Cache
	if enableCache {
		cache, err = GetValkeyInstane(ctx)
		if err != nil {
			t.Fatal(err)
		}

		config.Cache.Enabled = &enableCache
		config.Cache.Host = cache.Host
		config.Cache.Port = cache.Port
	}

	app := NewApp(ctx, config)
	ts := httptest.NewServer(app.Server.Handler)
	deps := Dependencies{Db: db, Cache: cache, App: app, TestServer: ts}
	err = deps.Db.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}

	return deps
}

func CreateAccessToken(t *testing.T, config config.AuthConfig, hours int, id *uuid.UUID, valid bool) string {
	now := time.Now()
	start := now.Add(-24 * time.Hour)
	expiresAt := now.Add(time.Duration(hours) * time.Hour)

	var sub string
	if id != nil {
		sub = id.String()
	} else {
		sub = ""
	}

	claims := handlers.JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    config.JwtIssuer,
			IssuedAt:  jwt.NewNumericDate(start),
			NotBefore: jwt.NewNumericDate(start),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	signedToken, err := token.SignedString(config.JwtEcdsaParsedKey)
	if err != nil {
		t.Fatal(err)
	}

	if !valid {
		tokenLen := len(signedToken)
		var n string
		lastChar := signedToken[tokenLen-1]
		if string(lastChar) != "a" {
			n = "a"
		} else {
			n = "b"
		}

		signedToken = signedToken[:tokenLen-1] + n
	}

	return signedToken
}
