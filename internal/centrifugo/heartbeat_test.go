package centrifugo_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/centrifugal/gocent/v3"
	"github.com/gopcua/opcua"
)

func TestPublishStatus(t *testing.T) {
	cases := []struct {
		name         string
		opcState     opcua.ConnState
		publishError bool
		expectError  bool
		expectData   string
	}{
		{
			name:         "PublishError",
			opcState:     opcua.Connected,
			publishError: true,
			expectError:  true,
			expectData:   "",
		},
		{
			name:         "OpcUaClosed",
			opcState:     opcua.Closed,
			publishError: false,
			expectError:  false,
			expectData:   `{"status":1,"description":"OPC-UA not connected"}`,
		},
		{
			name:         "OpcUaConnecting",
			opcState:     opcua.Connecting,
			publishError: false,
			expectError:  false,
			expectData:   `{"status":1,"description":"OPC-UA not connected"}`,
		},
		{
			name:         "OpcUaDisconnected",
			opcState:     opcua.Disconnected,
			publishError: false,
			expectError:  false,
			expectData:   `{"status":1,"description":"OPC-UA not connected"}`,
		},
		{
			name:         "OpcUaReconnecting",
			opcState:     opcua.Reconnecting,
			publishError: false,
			expectError:  false,
			expectData:   `{"status":1,"description":"OPC-UA not connected"}`,
		},
		{
			name:         "OpcUaConnected",
			opcState:     opcua.Connected,
			publishError: false,
			expectError:  false,
			expectData:   `{"status":0,"description":"Everything OK"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedStateProvider := &StateProviderMock{
				StateFunc: func() opcua.ConnState {
					return tc.opcState
				},
			}
			mockedPublisher := &PublisherMock{
				PublishFunc: func(ctx context.Context, channel string, data []byte, opts ...gocent.PublishOption) (gocent.PublishResult, error) {
					if tc.publishError {
						return gocent.PublishResult{}, testutils.ErrTesting
					}
					return gocent.PublishResult{}, nil
				},
			}

			err := PublishStatus(context.Background(), "testns", mockedStateProvider, mockedPublisher)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := len(mockedPublisher.PublishCalls()), 1; got != want {
				t.Errorf("Publish() call count: want %d, got %d", want, got)
			}
			if got, want := mockedPublisher.PublishCalls()[0].Channel, fmt.Sprintf("testns:%s", HeartbeatChannel); got != want {
				t.Errorf("Publish() `channel` argument: want %q, got %q", want, got)
			}
			if got, want := string(mockedPublisher.PublishCalls()[0].Data), tc.expectData; got != want {
				t.Errorf("Publish() `data` argument: want %q, got %q", want, got)
			}
		})
	}
}
