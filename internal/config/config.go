package config

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port              string `mapstructure:"port" validate:"required"`
	ListenAddr        string `mapstructure:"listenaddr"`
	Domain            string `mapstructure:"domain" validate:"required"`
	HttpsEnabled      bool   `mapstructure:"https_enabled"`      // For when the application server is behind a reverse proxy that handles TLS. If a certificate is provided and TLS is handled by the application server, it will always be true.
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

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver" validate:"required,oneof=postgres"`
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
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
	v.SetDefault("database.port", "5432")
	v.SetDefault("server.allow_login", false)
	v.SetDefault("server.allow_registration", false)
	v.SetDefault("server.allow_anonymous", false)

	v.SetDefault("server.auth.jwt_signing_method", "ES512")
	v.SetDefault("server.auth.jwt_issuer", "shurl")
}

func TrimConfigs(config Config) Config {
	config.Server.Port = strings.TrimSpace(config.Server.Port)
	config.Server.ListenAddr = strings.TrimSpace(config.Server.ListenAddr)
	config.Server.Auth.JwtIssuer = strings.TrimSpace(config.Server.Auth.JwtIssuer)
	config.Server.Auth.JwtKey = strings.TrimSpace(config.Server.Auth.JwtKey)
	config.Server.Auth.JwtSigningMethod = strings.TrimSpace(config.Server.Auth.JwtSigningMethod)

	config.Database.Driver = strings.TrimSpace(config.Database.Driver)
	config.Database.Host = strings.TrimSpace(config.Database.Host)
	config.Database.Port = strings.TrimSpace(config.Database.Port)
	config.Database.Name = strings.TrimSpace(config.Database.Name)
	config.Database.Username = strings.TrimSpace(config.Database.Username)
	config.Database.Password = strings.TrimSpace(config.Database.Password)

	return config
}

func LoadConfig(configFilePath string, ctx context.Context, logger utils.CustomJsonLogger) (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	SetDefaults(v)

	if configFilePath != "" {
		configFilePathInfo, err := os.Stat(configFilePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				logger.ErrorExit(ctx, "Config file provided cannot be found", "filepath", configFilePath)
			} else {
				logger.ErrorExit(ctx, "Error checking config file provided", "error", err.Error())
			}
		}
		if configFilePathInfo.IsDir() {
			logger.ErrorExit(ctx, "Config file provided is a directory, not a file", "filepath", configFilePath)
		}

		fullFileName := filepath.Base(configFilePath)
		fileExtension := strings.TrimPrefix(filepath.Ext(fullFileName), ".")

		if fileExtension == "" || !slices.Contains(AllowedConfigFileTypes, fileExtension) {
			errMessage := fmt.Sprintf("Unknown file extension '%s'", fileExtension)
			return nil, errors.New(errMessage)
		}

		v.SetConfigFile(configFilePath)
		v.SetConfigType(fileExtension)
		v.ReadInConfig()
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var config Config
	err := v.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	config = TrimConfigs(config)

	validate := validator.New()
	err = validate.Struct(&config)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(config.Server.Auth.JwtKey))
	if keyBlock == nil {
		return nil, errors.New("Could not parse ecdsa.PrivateKey")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	var ok bool
	config.Server.Auth.JwtEcdsaParsedKey, ok = key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("Could not parse ecdsa.PrivateKey")
	}

	return &config, nil
}
