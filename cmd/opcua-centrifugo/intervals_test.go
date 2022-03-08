package main_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/cmd/opcua-centrifugo"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
	"github.com/centrifugal/gocent/v3"
)

func TestChannelsInterval(t *testing.T) {
	cases := []struct {
		name            string
		channelsError   bool
		expectError     bool
		expectIntervals []time.Duration
	}{
		{
			name:            "ChannelsError",
			channelsError:   true,
			expectError:     true,
			expectIntervals: nil,
		},
		{
			name:            "Success",
			channelsError:   false,
			expectError:     false,
			expectIntervals: []time.Duration{2 * time.Second, 5 * time.Second},
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
							"ns:ch1@2000": {},
							"ch2":         {},
							"ns:ch3@5000": {},
						},
					}, nil
				},
			}

			in, err := ChannelIntervals(context.Background(), mockedCentrifugoChannels, "ns")

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Error(msg)
			}
			if got, want := in, tc.expectIntervals; !reflect.DeepEqual(got, want) {
				t.Errorf("returned intervals, want %#q, got %#q", want, got)
			}
		})
	}
}
