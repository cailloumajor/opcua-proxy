package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/joho/godotenv"
	"github.com/vrischmann/envconfig"
)

// Initializer describes methods to initialize the configuration
type Initializer interface {
	LoadEnvFile() error
	InitEnvConfig(*Config) error
}

// DefaultInitializer is the default configuration initializer
type DefaultInitializer struct{}

// LoadEnvFile loads environment variables from a .env file
func (*DefaultInitializer) LoadEnvFile() error {
	return godotenv.Load()
}

// InitEnvConfig initializes the configuration from environment variables
func (*DefaultInitializer) InitEnvConfig(c *Config) error {
	return envconfig.Init(c)
}

// Config holds the configuration of the application
type Config struct{}

// Init initializes and returns the application configuration
func Init(ci Initializer) (*Config, error) {
	cfg := &Config{}

	if err := ci.LoadEnvFile(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	if err := ci.InitEnvConfig(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	return cfg, nil
}
