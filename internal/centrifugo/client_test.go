package centrifugo_test

import (
	"context"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/centrifugal/gocent/v3"
)

func TestNewClient(t *testing.T) {
	c := NewClient("addr", "key")

	if c == nil {
		t.Error("unexpected nil client")
	}
}

func TestClientChannels(t *testing.T) {
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
			mockedClientProvider := &ClientProviderMock{
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
			c := &Client{mockedClientProvider}

			chs, err := c.Channels(context.Background(), "ns")

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Error(msg)
			}
			if got, want := len(chs), tc.expectIntervals; got != want {
				t.Errorf("returned intervals count, want %d, got %d", want, got)
			}
		})
	}
}

func TestClientHealth(t *testing.T) {
	cases := []struct {
		name            string
		infoError       bool
		expectedHealthy bool
		expectedMessage string
	}{
		{
			name:            "Unhealthy",
			infoError:       true,
			expectedHealthy: false,
			expectedMessage: testutils.ErrTesting.Error(),
		},
		{
			name:            "Healthy",
			infoError:       false,
			expectedHealthy: true,
			expectedMessage: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedClientProvider := &ClientProviderMock{
				InfoFunc: func(ctx context.Context) (gocent.InfoResult, error) {
					if tc.infoError {
						return gocent.InfoResult{}, testutils.ErrTesting
					}
					return gocent.InfoResult{}, nil
				},
			}
			c := &Client{mockedClientProvider}

			h, msg := c.Health(context.Background())

			if got, want := h, tc.expectedHealthy; got != want {
				t.Errorf("healthy status: want %v, got %v", want, got)
			}
			if got, want := msg, tc.expectedMessage; got != want {
				t.Errorf("health message: want %q, got %q", want, got)
			}
		})
	}
}
