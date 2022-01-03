package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/cailloumajor/opcua-centrifugo/internal/opcua"

	"github.com/joho/godotenv"
	"github.com/vrischmann/envconfig"
)

type initializer interface {
	loadEnvFile() error
	initEnvConfig(*Config) error
}

type defaultInitializer struct{}

func (defaultInitializer) loadEnvFile() error {
	return godotenv.Load()
}

func (defaultInitializer) initEnvConfig(c *Config) error {
	return envconfig.Init(c)
}

var di initializer

func init() {
	di = defaultInitializer{}
}

// Config holds the configuration of the application
type Config struct {
	Opcua opcua.Config
}

// Init initializes and returns the application configuration
func Init() (*Config, error) {
	cfg := &Config{}

	if err := di.loadEnvFile(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	if err := di.initEnvConfig(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	return cfg, nil
}
