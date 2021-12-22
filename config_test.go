package main

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type envFileLoaderMock struct {
	returnError error
}

func (e envFileLoaderMock) LoadEnvFile() error {
	return e.returnError
}

type envConfigInitializerMock struct {
	returnError error
}

func (e envConfigInitializerMock) InitEnvConfig(*Config) error {
	return e.returnError
}

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
			efl = envFileLoaderMock{tc.loadEnvFileError}
			eci = envConfigInitializerMock{tc.initEnvConfigError}
			_, err := InitConfig()
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

	expected := &Config{}

	for k, v := range envMap {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("error setting environment variable %v: %v", k, err)
		}
	}
	efl = envFileLoaderMock{}
	c, err := InitConfig()
	assert.NoError(err)
	assert.Equal(expected, c)
}
