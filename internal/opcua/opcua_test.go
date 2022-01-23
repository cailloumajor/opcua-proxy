package opcua_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/go-kit/log"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

var errTesting = errors.New("general error for testing")

func TestNewMonitorSuccess(t *testing.T) {
	mockedEndpointDescription := &ua.EndpointDescription{EndpointURL: "selectedEndpointURL"}
	mockedClient := &ClientMock{
		ConnectFunc: func(contextMoqParam context.Context) error {
			return nil
		},
	}
	mockedNodeMonitor := &NodeMonitorMock{
		SetErrorHandlerFunc: func(cb monitor.ErrHandler) {},
	}
	mockedNewMonitorDeps := &NewMonitorDepsMock{
		GetEndpointsFunc: func(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
			return []*ua.EndpointDescription{}, nil
		},
		SelectEndpointFunc: func(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
			return mockedEndpointDescription
		},
		NewClientFunc: func(endpoint string, opts ...opcua.Option) Client {
			return mockedClient
		},
		NewNodeMonitorFunc: func(client Client) (NodeMonitor, error) {
			return mockedNodeMonitor, nil
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
	m, err := NewMonitor(
		context.Background(),
		cfg,
		mockedNewMonitorDeps,
		mockedSecurityProvider,
		log.NewNopLogger(),
	)

	// Assertions about GetEndpoints
	if got, want := len(mockedNewMonitorDeps.GetEndpointsCalls()), 1; got != want {
		t.Errorf("GetEndpoints call count: want %d, got %d", want, got)
	}
	if got, want := mockedNewMonitorDeps.GetEndpointsCalls()[0].Endpoint, "serverURL"; got != want {
		t.Errorf("GetEndpoints endpoint argument: want %q, got %q", want, got)
	}
	if got, want := len(mockedNewMonitorDeps.GetEndpointsCalls()[0].Opts), 0; got != want {
		t.Errorf("GetEndpoints opts argument length: want %d, got %d", want, got)
	}
	// Assertions about SelectEndpoint
	if got, want := len(mockedNewMonitorDeps.SelectEndpointCalls()), 1; got != want {
		t.Errorf("SelectEndpoint call count: want %d, got %d", want, got)
	}
	if got, want := mockedNewMonitorDeps.SelectEndpointCalls()[0].Policy, "testpolicy"; got != want {
		t.Errorf("SelectEndpoint policy argument: want %q, got %q", want, got)
	}
	if got, want := mockedNewMonitorDeps.SelectEndpointCalls()[0].Mode, ua.MessageSecurityMode(42); got != want {
		t.Errorf("SelectEndpoint mode argument: want %q, got %q", want, got)
	}
	// Assertions about NewClient
	if got, want := len(mockedNewMonitorDeps.NewClientCalls()), 1; got != want {
		t.Errorf("NewClient call count: want %d, got %d", want, got)
	}
	if got, want := mockedNewMonitorDeps.NewClientCalls()[0].Endpoint, "selectedEndpointURL"; got != want {
		t.Errorf("NewClient endpoint argument: want %q, got %q", want, got)
	}
	if got, want := len(mockedNewMonitorDeps.NewClientCalls()[0].Opts), 8; got != want {
		t.Errorf("NewClient opts arguments count: want %d, got %d", want, got)
	}
	// Assertions about Client.Connect
	if got, want := len(mockedClient.ConnectCalls()), 1; got != want {
		t.Errorf("Client.Connect call count: want %d, got %d", want, got)
	}
	// Assertions about NewNodeMonitor
	if got, want := len(mockedNewMonitorDeps.NewNodeMonitorCalls()), 1; got != want {
		t.Errorf("NewNodeMonitor call count: want %d, got %d", want, got)
	}
	if got, want := mockedNewMonitorDeps.NewNodeMonitorCalls()[0].Client, mockedClient; got != want {
		t.Errorf("NewNodeMonitor client argument: want %+v, got %+v", want, got)
	}
	// Assertions about NodeMonitor.SetErrorHandler
	if got, want := len(mockedNodeMonitor.SetErrorHandlerCalls()), 1; got != want {
		t.Errorf("NodeMonitor.SetErrorHandler call count: want %d, got %d", want, got)
	}
	// Assertions about NewMonitor
	if m == nil {
		t.Errorf("NewMonitor return, got nil")
	}
	if got := err; got != nil {
		t.Errorf("NewMonitor error return: want nil, got %v", got)
	}
}

func TestNewMonitorError(t *testing.T) {
	cases := []struct {
		name                string
		getEndpointsError   bool
		selectEndpointNil   bool
		clientConnectError  bool
		newNodeMonitorError bool
	}{
		{
			name:                "GetEndpointsError",
			getEndpointsError:   true,
			selectEndpointNil:   false,
			clientConnectError:  false,
			newNodeMonitorError: false,
		},
		{
			name:                "SelectEndpointNil",
			getEndpointsError:   false,
			selectEndpointNil:   true,
			clientConnectError:  false,
			newNodeMonitorError: false,
		},
		{
			name:                "ClientConnectError",
			getEndpointsError:   false,
			selectEndpointNil:   false,
			clientConnectError:  true,
			newNodeMonitorError: false,
		},
		{
			name:                "NewNodeMonitorError",
			getEndpointsError:   false,
			selectEndpointNil:   false,
			clientConnectError:  false,
			newNodeMonitorError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedNewMonitorDeps := &NewMonitorDepsMock{
				GetEndpointsFunc: func(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
					if tc.getEndpointsError {
						return nil, errTesting
					}
					return []*ua.EndpointDescription{}, nil
				},
				SelectEndpointFunc: func(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
					if tc.selectEndpointNil {
						return nil
					}
					return &ua.EndpointDescription{}
				},
				NewClientFunc: func(endpoint string, opts ...opcua.Option) Client {
					return &ClientMock{
						ConnectFunc: func(contextMoqParam context.Context) error {
							if tc.clientConnectError {
								return errTesting
							}
							return nil
						},
					}
				},
				NewNodeMonitorFunc: func(client Client) (NodeMonitor, error) {
					if tc.newNodeMonitorError {
						return nil, errTesting
					}
					return &NodeMonitorMock{}, nil
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
			m, err := NewMonitor(
				context.Background(),
				&Config{},
				mockedNewMonitorDeps,
				mockedSecurityProvider,
				log.NewNopLogger(),
			)

			// Assertions about NewMonitor
			if got := m; got != nil {
				t.Errorf("NewMonitor return: want nil, got %+v", got)
			}
			if err == nil {
				t.Error("NewMonitor error return: want an error, got nil")
			}
		})
	}
}

func TestMonitorStop(t *testing.T) {
	mockedClient := &ClientMock{
		ConnectFunc: func(contextMoqParam context.Context) error {
			return nil
		},
		CloseFunc: func() error {
			return errTesting
		},
	}
	mockedNewMonitorDeps := &NewMonitorDepsMock{
		GetEndpointsFunc: func(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
			return []*ua.EndpointDescription{}, nil
		},
		SelectEndpointFunc: func(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
			return &ua.EndpointDescription{}
		},
		NewClientFunc: func(endpoint string, opts ...opcua.Option) Client {
			return mockedClient
		},
		NewNodeMonitorFunc: func(client Client) (NodeMonitor, error) {
			return &NodeMonitorMock{
				SetErrorHandlerFunc: func(cb monitor.ErrHandler) {},
			}, nil
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

	m, _ := NewMonitor(
		context.Background(),
		&Config{},
		mockedNewMonitorDeps,
		mockedSecurityProvider,
		log.NewNopLogger(),
	)
	var mockedSubscriptions [5]*SubscriptionMock
	for i := range mockedSubscriptions {
		mockedSubscription := &SubscriptionMock{
			UnsubscribeFunc: func(ctx context.Context) error {
				return errTesting
			},
		}
		mockedSubscriptions[i] = mockedSubscription
		m.AddSubscription(time.Duration(i+1)*time.Second, mockedSubscription)
	}

	errs := m.Stop(context.Background())

	if got, want := len(mockedClient.CloseCalls()), 1; got != want {
		t.Errorf("client.Close call count: want %d, got %d", want, got)
	}
	for _, v := range mockedSubscriptions {
		if got, want := len(v.UnsubscribeCalls()), 1; got != want {
			t.Errorf("Subscription.Unsubscribe call count: want %d, got %d", want, got)
		}
	}
	if got, want := len(errs), 6; got != want {
		t.Errorf("errors count: want %d, got %d", want, got)
	}
}
