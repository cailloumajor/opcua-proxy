package opcua_test

import (
	"context"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func TestNewClientSuccess(t *testing.T) {
	mockedEndpointDescription := &ua.EndpointDescription{EndpointURL: "selectedEndpointURL"}
	mockedRawClientProvider := &RawClientProviderMock{
		ConnectFunc: func(contextMoqParam context.Context) error {
			return nil
		},
		SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error) {
			return &opcua.Subscription{}, nil
		},
	}
	mockedClientExtDeps := &ClientExtDepsMock{
		GetEndpointsFunc: func(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
			return []*ua.EndpointDescription{}, nil
		},
		SelectEndpointFunc: func(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
			return mockedEndpointDescription
		},
		NewClientFunc: func(endpoint string, opts ...opcua.Option) RawClientProvider {
			return mockedRawClientProvider
		},
	}
	mockedSecurityProvider := &SecurityProviderMock{
		MessageSecurityModeFunc: func() ua.MessageSecurityMode {
			return ua.MessageSecurityMode(42)
		},
		PolicyFunc: func() string { return "testpolicy" },
		OptionsFunc: func(ep *ua.EndpointDescription) []opcua.Option {
			opts := make([]opcua.Option, 8)
			return opts
		},
	}

	cfg := &Config{
		ServerURL: "serverURL",
	}
	// Target of the test
	m, err := NewClient(
		context.Background(),
		cfg,
		mockedClientExtDeps,
		mockedSecurityProvider,
	)

	// Assertions about GetEndpoints
	if got, want := len(mockedClientExtDeps.GetEndpointsCalls()), 1; got != want {
		t.Errorf("GetEndpoints call count: want %d, got %d", want, got)
	}
	if got, want := mockedClientExtDeps.GetEndpointsCalls()[0].Endpoint, "serverURL"; got != want {
		t.Errorf("GetEndpoints endpoint argument: want %q, got %q", want, got)
	}
	if got, want := len(mockedClientExtDeps.GetEndpointsCalls()[0].Opts), 0; got != want {
		t.Errorf("GetEndpoints opts argument length: want %d, got %d", want, got)
	}
	// Assertions about SelectEndpoint
	if got, want := len(mockedClientExtDeps.SelectEndpointCalls()), 1; got != want {
		t.Errorf("SelectEndpoint call count: want %d, got %d", want, got)
	}
	if got, want := mockedClientExtDeps.SelectEndpointCalls()[0].Policy, "testpolicy"; got != want {
		t.Errorf("SelectEndpoint policy argument: want %q, got %q", want, got)
	}
	if got, want := mockedClientExtDeps.SelectEndpointCalls()[0].Mode, ua.MessageSecurityMode(42); got != want {
		t.Errorf("SelectEndpoint mode argument: want %q, got %q", want, got)
	}
	// Assertions about NewClient
	if got, want := len(mockedClientExtDeps.NewClientCalls()), 1; got != want {
		t.Errorf("NewClient call count: want %d, got %d", want, got)
	}
	if got, want := mockedClientExtDeps.NewClientCalls()[0].Endpoint, "selectedEndpointURL"; got != want {
		t.Errorf("NewClient endpoint argument: want %q, got %q", want, got)
	}
	if got, want := len(mockedClientExtDeps.NewClientCalls()[0].Opts), 8; got != want {
		t.Errorf("NewClient opts arguments count: want %d, got %d", want, got)
	}
	// Assertions about Client.Connect
	if got, want := len(mockedRawClientProvider.ConnectCalls()), 1; got != want {
		t.Errorf("Client.Connect call count: want %d, got %d", want, got)
	}
	// Assertions about NewClient
	if m == nil {
		t.Errorf("NewClient return, got nil")
	}
	if got := err; got != nil {
		t.Errorf("NewClient error return: want nil, got %v", got)
	}
}

func TestNewClientError(t *testing.T) {
	cases := []struct {
		name               string
		getEndpointsError  bool
		selectEndpointNil  bool
		clientConnectError bool
		subscribeError     bool
	}{
		{
			name:               "GetEndpointsError",
			getEndpointsError:  true,
			selectEndpointNil:  false,
			clientConnectError: false,
			subscribeError:     false,
		},
		{
			name:               "SelectEndpointNil",
			getEndpointsError:  false,
			selectEndpointNil:  true,
			clientConnectError: false,
			subscribeError:     false,
		},
		{
			name:               "ClientConnectError",
			getEndpointsError:  false,
			selectEndpointNil:  false,
			clientConnectError: true,
			subscribeError:     false,
		},
		{
			name:               "SubscribeError",
			getEndpointsError:  false,
			selectEndpointNil:  false,
			clientConnectError: false,
			subscribeError:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedRawClientProvider := &RawClientProviderMock{
				ConnectFunc: func(contextMoqParam context.Context) error {
					if tc.clientConnectError {
						return testutils.ErrTesting
					}
					return nil
				},
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error) {
					if tc.subscribeError {
						return nil, testutils.ErrTesting
					}
					return &opcua.Subscription{}, nil
				},
			}
			mockedClientExtDeps := &ClientExtDepsMock{
				GetEndpointsFunc: func(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
					if tc.getEndpointsError {
						return nil, testutils.ErrTesting
					}
					return []*ua.EndpointDescription{}, nil
				},
				SelectEndpointFunc: func(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
					if tc.selectEndpointNil {
						return nil
					}
					return &ua.EndpointDescription{}
				},
				NewClientFunc: func(endpoint string, opts ...opcua.Option) RawClientProvider {
					return mockedRawClientProvider
				},
			}
			mockedSecurityProvider := &SecurityProviderMock{
				MessageSecurityModeFunc: func() ua.MessageSecurityMode {
					return ua.MessageSecurityModeInvalid
				},
				PolicyFunc: func() string { return "" },
				OptionsFunc: func(ep *ua.EndpointDescription) []opcua.Option {
					return []opcua.Option{}
				},
			}

			// Target of the test
			_, err := NewClient(
				context.Background(),
				&Config{},
				mockedClientExtDeps,
				mockedSecurityProvider,
			)

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}

func TestNamespaceIndex(t *testing.T) {
	cases := []struct {
		name                string
		namespaceURI        string
		namespaceArrayError bool
		expectIndex         uint16
		expectError         bool
	}{
		{
			name:                "NamespaceArrayError",
			namespaceURI:        "ns0",
			namespaceArrayError: true,
			expectIndex:         0,
			expectError:         true,
		},
		{
			name:                "NamespaceNotFound",
			namespaceURI:        "null",
			namespaceArrayError: false,
			expectIndex:         0,
			expectError:         true,
		},
		{
			name:                "FirstNamespace",
			namespaceURI:        "ns0",
			namespaceArrayError: false,
			expectIndex:         0,
			expectError:         false,
		},
		{
			name:                "LastNamespace",
			namespaceURI:        "ns3",
			namespaceArrayError: false,
			expectIndex:         3,
			expectError:         false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedRawClientProvider := &RawClientProviderMock{
				ConnectFunc: func(contextMoqParam context.Context) error { return nil },
				NamespaceArrayWithContextFunc: func(ctx context.Context) ([]string, error) {
					if tc.namespaceArrayError {
						return nil, testutils.ErrTesting
					}
					return []string{"ns0", "ns1", "ns2", "ns3"}, nil
				},
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error) {
					return &opcua.Subscription{}, nil
				},
			}
			mockedClientExtDeps := &ClientExtDepsMock{
				GetEndpointsFunc: func(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
					return []*ua.EndpointDescription{}, nil
				},
				SelectEndpointFunc: func(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
					return &ua.EndpointDescription{}
				},
				NewClientFunc: func(endpoint string, opts ...opcua.Option) RawClientProvider {
					return mockedRawClientProvider
				},
			}
			mockedSecurityProvider := &SecurityProviderMock{
				MessageSecurityModeFunc: func() ua.MessageSecurityMode {
					return ua.MessageSecurityModeInvalid
				},
				PolicyFunc: func() string { return "" },
				OptionsFunc: func(ep *ua.EndpointDescription) []opcua.Option {
					return []opcua.Option{}
				},
			}
			c, _ := NewClient(context.Background(), &Config{}, mockedClientExtDeps, mockedSecurityProvider)

			idx, err := c.NamespaceIndex(context.Background(), tc.namespaceURI)

			if got, want := idx, tc.expectIndex; got != want {
				t.Errorf("index: want %d, got %d", want, got)
			}
			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}
