package main

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/joho/godotenv"
	"github.com/vrischmann/envconfig"
)

type envFileLoader interface {
	LoadEnvFile() error
}

type defEnvFileLoader struct{}

func (defEnvFileLoader) LoadEnvFile() error {
	return godotenv.Load()
}

type envConfigInitializer interface {
	InitEnvConfig(*Config) error
}

type defEnvConfigInitializer struct{}

func (defEnvConfigInitializer) InitEnvConfig(c *Config) error {
	return envconfig.Init(c)
}

// Config holds the configuration of the application
type Config struct{}

var (
	efl envFileLoader
	eci envConfigInitializer
)

func init() {
	efl = defEnvFileLoader{}
	eci = defEnvConfigInitializer{}
}

// InitConfig initializes and returns the application configuration
func InitConfig() (*Config, error) {
	cfg := &Config{}

	if err := efl.LoadEnvFile(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	if err := eci.InitEnvConfig(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	return cfg, nil
}
