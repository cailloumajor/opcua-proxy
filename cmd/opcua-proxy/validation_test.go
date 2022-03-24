package main_test

import (
	"testing"

	. "github.com/cailloumajor/opcua-proxy/cmd/opcua-proxy"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
)

func TestValidateCentrifugoAddress(t *testing.T) {
	cases := []struct {
		name        string
		address     string
		expectError bool
	}{
		{
			name:        "EmptyURL",
			address:     "",
			expectError: true,
		},
		{
			name:        "InvalidScheme",
			address:     "ftp://centrifugo.example.com:8000/api",
			expectError: true,
		},
		{
			name:        "GoodHttpURL",
			address:     "http://centrifugo.example.com:8000/api",
			expectError: false,
		},
		{
			name:        "GoodHttpsURL",
			address:     "https://centrifugo.example.com:8000/api",
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCentrifugoAddress(tc.address)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}
