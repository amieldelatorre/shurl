package internal

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/handlers"
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
	SERVER_DOMAIN = "localhost"
	DB_VERSION    = "20260304055040"
	DB_DRIVER     = "postgres"
	DB_NAME       = "shurl"
	DB_USERNAME   = "shurl"
	DB_PASSWORD   = "password"
	JWT_KEY       = `-----BEGIN PRIVATE KEY-----
MIHuAgEAMBAGByqGSM49AgEGBSuBBAAjBIHWMIHTAgEBBEIBxqUZyjGYLoZ12MOt
E7LMqwi4jlmni3JE6rEFHRYgMAxHpBZIzA1DFMaJUSvhHoG7IDEUuh4dYdJKcORT
crZz8nOhgYkDgYYABAGfTZFTug8rVyDng2JCCENWr9lnXoSETRk5p+3qi9Y7HAMM
JpBr7R1JRHprFqI08godS7mRE/ZuGnwNs0BdCnrxGgBlcSbEelp0GPLdkd+MGhsd
+5hecbTP6p0c9AeaZxa+TB0WnRg4d5Kojl5dgNmV9MUmUiItFA4jdUiHoZj3W6AC
oQ==
-----END PRIVATE KEY-----`
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

	var cache Cache
	if enableCache {
		cache, err = GetValkeyInstane(ctx)
		if err != nil {
			t.Fatal(err)
		}

		t.Setenv("CACHE_ENABLED", "true")
		t.Setenv("CACHE_HOST", cache.Host)
		t.Setenv("CACHE_PORT", cache.Port)
	}

	t.Setenv("SERVER_DOMAIN", SERVER_DOMAIN)
	t.Setenv("DATABASE_RUN_MIGRATIONS", "true")
	t.Setenv("DATABASE_DRIVER", DB_DRIVER)
	t.Setenv("DATABASE_HOST", db.Host)
	t.Setenv("DATABASE_PORT", db.Port)
	t.Setenv("DATABASE_NAME", DB_NAME)
	t.Setenv("DATABASE_USERNAME", DB_USERNAME)
	t.Setenv("DATABASE_PASSWORD", DB_PASSWORD)
	t.Setenv("SERVER_AUTH_JWT_KEY", JWT_KEY)

	app := NewApp("")
	ts := httptest.NewServer(app.Server.Handler)
	deps := Dependencies{Db: db, Cache: cache, App: app, TestServer: ts}
	err = deps.Db.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}

	return deps
}

func CreateAccessToken(t *testing.T, config config.AuthConfig, hours int) string {
	now := time.Now()
	start := now.Add(-24 * time.Hour)
	expiresAt := now.Add(time.Duration(hours) * time.Hour)
	claims := handlers.JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.Nil.String(),
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

	return signedToken
}
