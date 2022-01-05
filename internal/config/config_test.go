package config_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/cailloumajor/opcua-centrifugo/internal/config"
)

var errTesting = errors.New("general error for testing")

// Tests the error return of InitConfig
func TestInit(t *testing.T) {
	tests := []struct {
		name               string
		loadEnvFileError   error
		initEnvConfigError error
		expectError        bool
	}{
		{"env file loading error", errTesting, nil, true},
		{"missing env file", fs.ErrNotExist, nil, false},
		{"env config loading error", nil, errTesting, true},
		{"no error", nil, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config.MockInitializer(tc.loadEnvFileError, tc.initEnvConfigError)
			_, err := config.Init()
			if tc.expectError && err == nil {
				t.Error("want an error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Error("want nil, got an error")
			}
		})
	}
}
