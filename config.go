package main

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/joho/godotenv"
	"github.com/vrischmann/envconfig"
)

type configHandler interface {
	LoadEnvFile() error
	InitEnvConfig(*Config) error
}

type defConfigHandler struct{}

func (defConfigHandler) LoadEnvFile() error {
	return godotenv.Load()
}

func (defConfigHandler) InitEnvConfig(c *Config) error {
	return envconfig.Init(c)
}

// Config holds the configuration of the application
type Config struct{}

var (
	ch configHandler
)

func init() {
	ch = defConfigHandler{}
}

// InitConfig initializes and returns the application configuration
func InitConfig() (*Config, error) {
	cfg := &Config{}

	if err := ch.LoadEnvFile(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	if err := ch.InitEnvConfig(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	return cfg, nil
}
