package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port" validate:"required"`
	ListenAddr   string `mapstructure:"listenaddr"`
	Domain       string `mapstructure:"domain" validate:"required"`
	HttpsEnabled bool   `mapstructure:"https_enabled"`
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
	v.SetDefault("server.https_enabled", true)
	v.SetDefault("database.port", "5432")
}

func TrimConfigs(config Config) Config {
	config.Server.Port = strings.TrimSpace(config.Server.Port)
	config.Server.ListenAddr = strings.TrimSpace(config.Server.ListenAddr)

	config.Database.Driver = strings.TrimSpace(config.Database.Driver)
	config.Database.Host = strings.TrimSpace(config.Database.Host)
	config.Database.Port = strings.TrimSpace(config.Database.Port)
	config.Database.Name = strings.TrimSpace(config.Database.Name)
	config.Database.Username = strings.TrimSpace(config.Database.Username)
	config.Database.Password = strings.TrimSpace(config.Database.Password)

	return config
}

func LoadConfig(configFilePath string) (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	SetDefaults(v)

	if configFilePath != "" {
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

	validate := validator.New()
	err = validate.Struct(&config)
	if err != nil {
		return nil, err
	}

	config = TrimConfigs(config)

	return &config, nil
}
