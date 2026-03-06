package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/spf13/viper"
)

type Config struct {
	Server                      ServerConfig                `mapstructure:"server"`
	Database                    DatabaseConfig              `mapstructure:"database"`
	IdempotencyKeyCleanupWorker IdempotencyKeyCleanupWorker `mapstructure:"idempotency_key_cleanup_worker"`
	ShortUrlCleanupWorker       ShortUrlCleanupWorker       `mapstructure:"short_url_cleanup_worker"`
	Cache                       CacheConfig                 `mapstructure:"cache"`
	Log                         LogConfig                   `mapstructure:"log"`
}

type ServerConfig struct {
	Port              string `mapstructure:"port" validate:"required"`
	ListenAddr        string `mapstructure:"listenaddr"`
	Domain            string `mapstructure:"domain" validate:"required"`
	HttpsEnabled      bool   `mapstructure:"https_enabled"`      // For when the application server is behind a reverse proxy that handles TLS. If a certificate is provided and TLS is handled by the application server, it will always be true.
	AppendPort        bool   `mapstructure:"append_port"`        // If the server port should be appended to the domain
	AllowLogin        bool   `mapstructure:"allow_login"`        // Allow login, by default only authenticated users are allowed to create urls
	AllowRegistration bool   `mapstructure:"allow_registration"` // Allow user registration, this also needs `server.allow_login` to be true in order to take effect
	AllowAnonymous    bool   `mapstructure:"allow_anonymous"`    // Allow anonymous link creation

	// TODO: Make this required only if allow login is true. For now, it is always required
	Auth AuthConfig `mapstructure:"auth"`
}

type AuthConfig struct {
	JwtSigningMethod string `mapstructure:"jwt_signing_method" validate:"required,oneof=ES512"`
	JwtKey           string `mapstructure:"jwt_key" validate:"required"`
	JwtIssuer        string `mapstructure:"jwt_issuer" validate:"required"`
	// TODO: Make it possible to read from a file that is passed in

	JwtEcdsaParsedKey *ecdsa.PrivateKey `mapstructure:"-" validate:"-"`
}

type IdempotencyKeyCleanupWorker struct {
	IntervalSeconds int  `mapstructure:"interval_seconds" validate:"required,min=300,max=21600"`
	ErrorsFatal     bool `mapstructure:"errors_fatal" validate:"required"`
}

type ShortUrlCleanupWorker struct {
	IntervalSeconds int  `mapstructure:"interval_seconds" validate:"required,min=300,max=21600"`
	ErrorsFatal     bool `mapstructure:"errors_fatal" validate:"required"`
}

type DatabaseConfig struct {
	RunMigrations *bool  `mapstructure:"run_migrations" validate:"required"`
	Driver        string `mapstructure:"driver" validate:"required,oneof=postgres"`
	Host          string `mapstructure:"host" validate:"required"`
	Port          string `mapstructure:"port" validate:"required"`
	Name          string `mapstructure:"name" validate:"required"`
	Username      string `mapstructure:"username" validate:"required"`
	Password      string `mapstructure:"password" validate:"required"`
}

type CacheConfig struct {
	Enabled *bool  `mapstructure:"enabled" validate:"required"`
	Driver  string `mapstructure:"driver" validate:"required_if=Enabled true,oneof=valkey"`
	Host    string `mapstructure:"host" validate:"required_if=Enabled true"`
	Port    string `mapstructure:"port" validate:"required_if=Enabled true"`
}

type LogConfig struct {
	Level     string     `mapstructure:"level" validate:"required,loglevelvalidator"`
	SlogLevel slog.Level `mapstructure:"-" validate:"-"`
}

var (
	AllowedConfigFileTypes = []string{
		"env",
		"ini",
		"toml",
		"yaml",
		"json",
	}
)

func SetDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.https_enabled", false)
	v.SetDefault("server.append_port", true)
	v.SetDefault("server.allow_login", false)
	v.SetDefault("server.allow_registration", false)
	v.SetDefault("server.allow_anonymous", false)
	v.SetDefault("server.auth.jwt_signing_method", "ES512")
	v.SetDefault("server.auth.jwt_issuer", "shurl")

	v.SetDefault("idempotency_key_cleanup_worker.interval_seconds", 600)
	v.SetDefault("idempotency_key_cleanup_worker.errors_fatal", true)

	v.SetDefault("short_url_cleanup_worker.interval_seconds", 600)
	v.SetDefault("short_url_cleanup_worker.errors_fatal", true)

	v.SetDefault("database.run_migrations", true)
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.port", "5432")

	v.SetDefault("cache.enabled", false)
	v.SetDefault("cache.driver", "valkey")
	v.SetDefault("cache.port", "6379")

	v.SetDefault("log.level", "INFO")
}

func TrimConfigs(config Config) Config {
	config.Server.Port = strings.TrimSpace(config.Server.Port)
	config.Server.ListenAddr = strings.TrimSpace(config.Server.ListenAddr)
	config.Server.Domain = strings.TrimSpace(config.Server.Domain)

	config.Server.Auth.JwtIssuer = strings.TrimSpace(config.Server.Auth.JwtIssuer)
	config.Server.Auth.JwtKey = strings.TrimSpace(config.Server.Auth.JwtKey)
	config.Server.Auth.JwtSigningMethod = strings.TrimSpace(config.Server.Auth.JwtSigningMethod)

	config.Database.Driver = strings.TrimSpace(config.Database.Driver)
	config.Database.Host = strings.TrimSpace(config.Database.Host)
	config.Database.Port = strings.TrimSpace(config.Database.Port)
	config.Database.Name = strings.TrimSpace(config.Database.Name)
	config.Database.Username = strings.TrimSpace(config.Database.Username)
	config.Database.Password = strings.TrimSpace(config.Database.Password)

	config.Cache.Driver = strings.TrimSpace(config.Cache.Driver)
	config.Cache.Host = strings.TrimSpace(config.Cache.Host)
	config.Cache.Port = strings.TrimSpace(config.Cache.Port)

	config.Log.Level = strings.TrimSpace(config.Log.Level)

	return config
}

func LoadConfig(configFilePath string) (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	SetDefaults(v)

	if configFilePath != "" {
		configFilePathInfo, err := os.Stat(configFilePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("Config file provided cannot be found '%s'", configFilePath)
			} else {
				return nil, err
			}
		}
		if configFilePathInfo.IsDir() {
			return nil, fmt.Errorf("Config file provided is a directory, not a file'%s'", configFilePath)
		}

		fullFileName := filepath.Base(configFilePath)
		fileExtension := strings.TrimPrefix(filepath.Ext(fullFileName), ".")

		if fileExtension == "" || !slices.Contains(AllowedConfigFileTypes, fileExtension) {
			errMessage := fmt.Sprintf("Unknown file extension '%s'", fileExtension)
			return nil, errors.New(errMessage)
		}

		v.SetConfigFile(configFilePath)
		v.SetConfigType(fileExtension)
		err = v.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var config Config
	err := v.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	config = TrimConfigs(config)

	validate, err := utils.GetValidator()
	if err != nil {
		return nil, err
	}

	err = validate.Struct(&config)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(config.Server.Auth.JwtKey))
	if keyBlock == nil {
		return nil, errors.New("could not parse ecdsa private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	var ok bool
	config.Server.Auth.JwtEcdsaParsedKey, ok = key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("could not parse ecdsa private key")
	}

	err = config.Log.SlogLevel.UnmarshalText([]byte(config.Log.Level))
	if err != nil {
		return nil, err
	}

	return &config, nil
}
