package main_test

import (
	"context"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/cmd/opcua-proxy"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/centrifugal/gocent/v3"
)

func TestChannels(t *testing.T) {
	cases := []struct {
		name            string
		channelsError   bool
		expectError     bool
		expectIntervals int
	}{
		{
			name:            "ChannelsError",
			channelsError:   true,
			expectError:     true,
			expectIntervals: 0,
		},
		{
			name:            "Success",
			channelsError:   false,
			expectError:     false,
			expectIntervals: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedCentrifugoChannels := &CentrifugoChannelsMock{
				ChannelsFunc: func(ctx context.Context, opts ...gocent.ChannelsOption) (gocent.ChannelsResult, error) {
					options := &gocent.ChannelsOptions{}
					for _, opt := range opts {
						opt(options)
					}
					if got, want := options.Pattern, "ns:*"; got != want {
						t.Errorf("Channels call Pattern option: want %q, got %q", want, got)
					}
					if tc.channelsError {
						return gocent.ChannelsResult{}, testutils.ErrTesting
					}
					return gocent.ChannelsResult{
						Channels: map[string]gocent.ChannelInfo{
							"ch1": {}, "ch2": {}, "ch3": {},
						},
					}, nil
				},
			}

			chs, err := Channels(context.Background(), mockedCentrifugoChannels, "ns")

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Error(msg)
			}
			if got, want := len(chs), tc.expectIntervals; got != want {
				t.Errorf("returned intervals count, want %d, got %d", want, got)
			}
		})
	}
}
