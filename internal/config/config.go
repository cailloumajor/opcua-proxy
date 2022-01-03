package config

import (
	"errors"
	"fmt"
	"io/fs"

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

// OpcUaConfig holds the OPC-UA part of the configuration
type OpcUaConfig struct {
	ServerURL string
	User      string `envconfig:"optional"`
	Password  string `envconfig:"optional"`
	CertFile  string `envconfig:"optional"`
	KeyFile   string `envconfig:"optional"`
}

// Config holds the configuration of the application
type Config struct {
	Opcua OpcUaConfig
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
