package opcua_test

import (
	"context"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type ClientOpts struct {
	user     string
	pass     string
	certFile string
	keyFile  string
}

func TestNewClient(t *testing.T) {
	cases := []struct {
		name                        string
		config                      Config
		getEndpointsError           bool
		selectEndpointNil           bool
		expectedMessageSecurityMode ua.MessageSecurityMode
		expectedPolicy              string
		expectClientOpts            ClientOpts
		expectOptsCount             int
		expectError                 bool
	}{
		{
			name:                        "GetEndpointsError",
			config:                      Config{},
			getEndpointsError:           true,
			selectEndpointNil:           false,
			expectedMessageSecurityMode: ua.MessageSecurityModeInvalid,
			expectedPolicy:              "",
			expectClientOpts:            ClientOpts{},
			expectOptsCount:             0,
			expectError:                 true,
		},
		{
			name:                        "SelectEndpointNil",
			config:                      Config{},
			getEndpointsError:           false,
			selectEndpointNil:           true,
			expectedMessageSecurityMode: ua.MessageSecurityModeInvalid,
			expectedPolicy:              "",
			expectClientOpts:            ClientOpts{},
			expectOptsCount:             0,
			expectError:                 true,
		},
		{
			name:                        "SuccessNoAuthNoEncryption",
			config:                      Config{},
			getEndpointsError:           false,
			selectEndpointNil:           false,
			expectedMessageSecurityMode: ua.MessageSecurityModeNone,
			expectedPolicy:              "None",
			expectClientOpts:            ClientOpts{},
			expectOptsCount:             1,
			expectError:                 false,
		},
		{
			name:                        "SuccessAuthNoEncryption",
			config:                      Config{User: "user", Password: "pass"},
			getEndpointsError:           false,
			selectEndpointNil:           false,
			expectedMessageSecurityMode: ua.MessageSecurityModeNone,
			expectedPolicy:              "None",
			expectClientOpts:            ClientOpts{user: "user", pass: "pass"},
			expectOptsCount:             2,
			expectError:                 false,
		},
		{
			name:                        "SuccessNoAuthEncryption",
			config:                      Config{CertFile: "certf.ile", KeyFile: "keyf.ile"},
			getEndpointsError:           false,
			selectEndpointNil:           false,
			expectedMessageSecurityMode: ua.MessageSecurityModeSignAndEncrypt,
			expectedPolicy:              "Basic256Sha256",
			expectClientOpts:            ClientOpts{certFile: "certf.ile", keyFile: "keyf.ile"},
			expectOptsCount:             3,
			expectError:                 false,
		},
		{
			name:                        "SuccessAuthEncryption",
			config:                      Config{User: "user", Password: "pass", CertFile: "certf.ile", KeyFile: "keyf.ile"},
			getEndpointsError:           false,
			selectEndpointNil:           false,
			expectedMessageSecurityMode: ua.MessageSecurityModeSignAndEncrypt,
			expectedPolicy:              "Basic256Sha256",
			expectClientOpts:            ClientOpts{user: "user", pass: "pass", certFile: "certf.ile", keyFile: "keyf.ile"},
			expectOptsCount:             4,
			expectError:                 false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var clientOpts ClientOpts
			ep := &ua.EndpointDescription{
				EndpointURL: "selectedEndpointURL",
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
					return ep
				},
				NewClientFunc: func(endpoint string, opts ...opcua.Option) *opcua.Client {
					return &opcua.Client{}
				},
				AuthUsernameFunc: func(user, pass string) opcua.Option {
					clientOpts.user = user
					clientOpts.pass = pass
					return func(c *opcua.Config) {}
				},
				CertificateFileFunc: func(filename string) opcua.Option {
					clientOpts.certFile = filename
					return func(c *opcua.Config) {}
				},
				PrivateKeyFileFunc: func(filename string) opcua.Option {
					clientOpts.keyFile = filename
					return func(c *opcua.Config) {}
				},
				SecurityFromEndpointFunc: func(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
					return func(c *opcua.Config) {}
				},
			}

			// Target of the test
			tc.config.ServerURL = "testServerURL"
			_, err := NewClient(context.Background(), &tc.config, mockedClientExtDeps)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if tc.expectError {
				return
			}
			// Assertions about GetEndpoints
			if got, want := mockedClientExtDeps.GetEndpointsCalls()[0].Endpoint, "testServerURL"; got != want {
				t.Errorf("GetEndpoints endpoint argument: want %q, got %q", want, got)
			}
			// Assertions about SelectEndpoint
			if got, want := mockedClientExtDeps.SelectEndpointCalls()[0].Policy, tc.expectedPolicy; got != want {
				t.Errorf("SelectEndpoint policy argument: want %q, got %q", want, got)
			}
			if got, want := mockedClientExtDeps.SelectEndpointCalls()[0].Mode, tc.expectedMessageSecurityMode; got != want {
				t.Errorf("SelectEndpoint mode argument: want %q, got %q", want, got)
			}
			// Assertions about SecurityFromEndpoint
			if got, want := mockedClientExtDeps.SecurityFromEndpointCalls()[0].Ep, ep; got != want {
				t.Errorf("SecurityFromEndpoint ep argument: want %v, got %v", want, got)
			}
			// Assertions about client options
			if got, want := clientOpts, tc.expectClientOpts; got != want {
				t.Errorf("Client options: want %v, got %v", want, got)
			}
			// Assertions about NewClient
			if got, want := mockedClientExtDeps.NewClientCalls()[0].Endpoint, "selectedEndpointURL"; got != want {
				t.Errorf("NewClient endpoint argument: want %q, got %q", want, got)
			}
			if got, want := len(mockedClientExtDeps.NewClientCalls()[0].Opts), tc.expectOptsCount; got != want {
				t.Errorf("NewClient opts arguments count: want %d, got %d", want, got)
			}
		})
	}
}
