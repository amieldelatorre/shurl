package test

import (
	"context"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
)

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
)

type Pg struct {
	Container *tcpostgres.PostgresContainer
	Host      string
	Port      string
}

type Valkey struct {
	Container *tcvalkey.ValkeyContainer
	Host      string
	Port      string
}

func GetValkeyInstane(ctx context.Context) (Valkey, error) {
	valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:9.0.3")
	if err != nil {
		return Valkey{}, err
	}

	host, err := valkeyContainer.Host(ctx)
	if err != nil {
		return Valkey{}, err
	}

	port, err := valkeyContainer.MappedPort(ctx, "6379")
	if err != nil {
		return Valkey{}, err
	}

	return Valkey{Container: valkeyContainer, Host: host, Port: port.Port()}, nil
}

func GetPostgreSqlInstace(ctx context.Context) (Pg, error) {
	postgresContainer, err := tcpostgres.Run(ctx,
		"postgres:18.3",
		tcpostgres.BasicWaitStrategies(),
		tcpostgres.WithDatabase(DB_NAME),
		tcpostgres.WithUsername(DB_USERNAME),
		tcpostgres.WithPassword(DB_PASSWORD),
	)
	if err != nil {
		return Pg{}, err
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return Pg{}, err
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return Pg{}, err
	}

	return Pg{Container: postgresContainer, Host: host, Port: port.Port()}, nil
}
