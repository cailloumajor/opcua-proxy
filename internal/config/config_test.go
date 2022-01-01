package config_test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/cailloumajor/opcua-centrifugo/internal/config"

	"github.com/stretchr/testify/assert"
)

// Tests the error return of InitConfig
func TestInitConfigError(t *testing.T) {
	tests := []struct {
		name               string
		loadEnvFileError   error
		initEnvConfigError error
		expectError        bool
	}{
		{"env file loading error", assert.AnError, nil, true},
		{"missing env file", fs.ErrNotExist, nil, false},
		{"env config loading error", nil, assert.AnError, true},
		{"no error", nil, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			config.MockInitializer(tc.loadEnvFileError, tc.initEnvConfigError)
			_, err := config.Init()
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestInitConfigSuccess(t *testing.T) {
	assert := assert.New(t)

	envMap := map[string]string{}

	expected := &config.Config{}

	for k, v := range envMap {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("error setting environment variable %v: %v", k, err)
		}
	}
	config.MockInitializer(nil, nil)
	c, err := config.Init()
	assert.NoError(err)
	assert.Equal(expected, c)
}
