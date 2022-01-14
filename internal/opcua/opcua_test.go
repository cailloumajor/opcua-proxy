package opcua_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	. "github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/go-kit/log"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

var errTesting = errors.New("general error for testing")

type authUserNameCall struct {
	calls uint
	user  string
	pass  string
}

type certificateFileCall struct {
	calls    uint
	filename string
}

type privateKeyFileCall struct {
	calls    uint
	filename string
}

func TestNewMonitorSuccess(t *testing.T) {
	cases := []struct {
		name                       string
		config                     Config
		expectSelectEndpointPolicy string
		expectSelectEndpointMode   ua.MessageSecurityMode
		expectAuthUserNameCalls    authUserNameCall
		expectCertificateFileCalls certificateFileCall
		expectPrivateKeyFileCalls  privateKeyFileCall
		expectUserTokenType        ua.UserTokenType
		expectNewClientOptsCount   int
	}{
		{
			name:                       "SuccessWithoutAuthWithoutEncryption",
			config:                     Config{},
			expectSelectEndpointPolicy: "None",
			expectSelectEndpointMode:   ua.MessageSecurityModeNone,
			expectAuthUserNameCalls:    authUserNameCall{},
			expectCertificateFileCalls: certificateFileCall{},
			expectPrivateKeyFileCalls:  privateKeyFileCall{},
			expectUserTokenType:        ua.UserTokenTypeAnonymous,
			expectNewClientOptsCount:   1,
		},
		{
			name:                       "SuccessWithAuthWithoutEncryption",
			config:                     Config{User: "user1", Password: "pass1"},
			expectSelectEndpointPolicy: "None",
			expectSelectEndpointMode:   ua.MessageSecurityModeNone,
			expectAuthUserNameCalls:    authUserNameCall{calls: 1, user: "user1", pass: "pass1"},
			expectCertificateFileCalls: certificateFileCall{},
			expectPrivateKeyFileCalls:  privateKeyFileCall{},
			expectUserTokenType:        ua.UserTokenTypeUserName,
			expectNewClientOptsCount:   2,
		},
		{
			name:                       "SuccessWithoutAuthWithEncryption",
			config:                     Config{CertFile: "cert1", KeyFile: "key1"},
			expectSelectEndpointPolicy: "Basic256Sha256",
			expectSelectEndpointMode:   ua.MessageSecurityModeSignAndEncrypt,
			expectAuthUserNameCalls:    authUserNameCall{},
			expectCertificateFileCalls: certificateFileCall{calls: 1, filename: "cert1"},
			expectPrivateKeyFileCalls:  privateKeyFileCall{calls: 1, filename: "key1"},
			expectUserTokenType:        ua.UserTokenTypeAnonymous,
			expectNewClientOptsCount:   3,
		},
		{
			name:                       "SuccessWithAuthWithEncryption",
			config:                     Config{User: "user2", Password: "pass2", CertFile: "cert2", KeyFile: "key2"},
			expectSelectEndpointPolicy: "Basic256Sha256",
			expectSelectEndpointMode:   ua.MessageSecurityModeSignAndEncrypt,
			expectAuthUserNameCalls:    authUserNameCall{calls: 1, user: "user2", pass: "pass2"},
			expectCertificateFileCalls: certificateFileCall{calls: 1, filename: "cert2"},
			expectPrivateKeyFileCalls:  privateKeyFileCall{calls: 1, filename: "key2"},
			expectUserTokenType:        ua.UserTokenTypeUserName,
			expectNewClientOptsCount:   4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				authUserNameCalls    authUserNameCall
				certificateFileCalls certificateFileCall
				privateKeyFileCalls  privateKeyFileCall
			)
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
				AuthUsernameFunc: func(user, pass string) opcua.Option {
					authUserNameCalls.calls++
					authUserNameCalls.user = user
					authUserNameCalls.pass = pass
					return func(c *opcua.Config) {}
				},
				CertificateFileFunc: func(filename string) opcua.Option {
					certificateFileCalls.calls++
					certificateFileCalls.filename = filename
					return func(c *opcua.Config) {}
				},
				PrivateKeyFileFunc: func(filename string) opcua.Option {
					privateKeyFileCalls.calls++
					privateKeyFileCalls.filename = filename
					return func(c *opcua.Config) {}
				},
				SecurityFromEndpointFunc: func(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
					return func(c *opcua.Config) {}
				},
				NewClientFunc: func(endpoint string, opts ...opcua.Option) Client {
					return mockedClient
				},
				NewNodeMonitorFunc: func(client Client) (NodeMonitor, error) {
					return mockedNodeMonitor, nil
				},
			}

			tc.config.ServerURL = "serverURL"

			// Target of the test
			m, err := NewMonitor(context.Background(), &tc.config, mockedNewMonitorDeps, log.NewNopLogger())

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
			if got, want := mockedNewMonitorDeps.SelectEndpointCalls()[0].Policy, tc.expectSelectEndpointPolicy; got != want {
				t.Errorf("SelectEndpoint policy argument: want %q, got %q", want, got)
			}
			if got, want := mockedNewMonitorDeps.SelectEndpointCalls()[0].Mode, tc.expectSelectEndpointMode; got != want {
				t.Errorf("SelectEndpoint mode argument: want %q, got %q", want, got)
			}
			// Assertions about AuthUserName
			if got, want := authUserNameCalls, tc.expectAuthUserNameCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("AuthUserName calls: want %#v, got %#v", want, got)
			}
			// Assertions about CertificateFile
			if got, want := certificateFileCalls, tc.expectCertificateFileCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("CertificateFile calls: want %#v, got %#v", want, got)
			}
			// Assertions about PrivateKeyFile
			if got, want := privateKeyFileCalls, tc.expectPrivateKeyFileCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("PrivateKeyFile calls: want %#v, got %#v", want, got)
			}
			// Assertions about SecurityFromEndpoint
			if got, want := len(mockedNewMonitorDeps.SecurityFromEndpointCalls()), 1; !reflect.DeepEqual(got, want) {
				t.Errorf("SecurityFromEndpoint call count: want %d, got %d", want, got)
			}
			if got, want := mockedNewMonitorDeps.SecurityFromEndpointCalls()[0].Ep, mockedEndpointDescription; got != want {
				t.Errorf("SecurityFromEndpoint ep argument: want %#v, got %#v", want, got)
			}
			if got, want := mockedNewMonitorDeps.SecurityFromEndpointCalls()[0].AuthType, tc.expectUserTokenType; got != want {
				t.Errorf("SecurityFromEndpoint authType argument: want %q, got %q", want, got)
			}
			// Assertions about NewClient
			if got, want := len(mockedNewMonitorDeps.NewClientCalls()), 1; got != want {
				t.Errorf("NewClient call count: want %d, got %d", want, got)
			}
			if got, want := mockedNewMonitorDeps.NewClientCalls()[0].Endpoint, "selectedEndpointURL"; got != want {
				t.Errorf("NewClient endpoint argument: want %q, got %q", want, got)
			}
			if got, want := len(mockedNewMonitorDeps.NewClientCalls()[0].Opts), tc.expectNewClientOptsCount; got != want {
				t.Errorf("NewClient opts arguments count: want %+v, got %+v", want, got)
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
		})
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
			mockedClient := &ClientMock{
				ConnectFunc: func(contextMoqParam context.Context) error {
					if tc.clientConnectError {
						return errTesting
					}
					return nil
				},
			}
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
				SecurityFromEndpointFunc: func(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
					return func(c *opcua.Config) {}
				},
				NewClientFunc: func(endpoint string, opts ...opcua.Option) Client {
					return mockedClient
				},
				NewNodeMonitorFunc: func(client Client) (NodeMonitor, error) {
					if tc.newNodeMonitorError {
						return nil, errTesting
					}
					return &NodeMonitorMock{}, nil
				},
			}

			// Target of the test
			m, err := NewMonitor(context.Background(), &Config{}, mockedNewMonitorDeps, log.NewNopLogger())

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
	const ExpectedErrors = 1

	mockedClient := &ClientMock{
		CloseFunc: func() error {
			return errTesting
		},
	}
	mockedNodeMonitor := &NodeMonitorMock{}

	m := NewTestMonitor(mockedClient, mockedNodeMonitor)
	errs := m.Stop(context.Background())

	if got, want := len(mockedClient.CloseCalls()), 1; got != want {
		t.Errorf("client.Close call count: want %d, got %d", want, got)
	}
	if got, want := len(errs), ExpectedErrors; got != want {
		t.Errorf("errors count: want %d, got %d", want, got)
	}
}
