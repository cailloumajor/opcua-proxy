package config_test

import (
	"io/fs"
	"testing"

	"github.com/cailloumajor/opcua-centrifugo/internal/config"

	"github.com/stretchr/testify/assert"
)

// Tests the error return of InitConfig
func TestInit(t *testing.T) {
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
