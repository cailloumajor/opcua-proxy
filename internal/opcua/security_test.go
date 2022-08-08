package opcua_test

import (
	"reflect"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func TestNewSecurityFailure(t *testing.T) {
	mockedSecurityExtDeps := &SecurityExtDepsMock{}

	cases := []struct {
		name   string
		config Config
	}{
		{
			name:   "PasswordWithoutUser",
			config: Config{Password: "pass"},
		},
		{
			name:   "CertFileWithoutKeyFile",
			config: Config{CertFile: "certf.ile"},
		},
		{
			name:   "KeyFileWithoutCertFile",
			config: Config{KeyFile: "keyf.ile"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSecurity(&tc.config, mockedSecurityExtDeps)

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf("NewSecurity(): %s", msg)
			}
		})
	}
}

type AuthUserNameCall struct {
	user string
	pass string
}

type CertificateFileCall struct {
	filename string
}

type PrivateKeyFileCall struct {
	filename string
}

func TestSecuritySuccess(t *testing.T) {
	cases := []struct {
		name                       string
		config                     Config
		expectAuthUserNameCalls    []AuthUserNameCall
		expectCertificateFileCalls []CertificateFileCall
		expectPrivateKeyFileCalls  []PrivateKeyFileCall
		expectUserTokenType        ua.UserTokenType
		expectMessageSecurityMode  ua.MessageSecurityMode
		expectPolicy               string
		expectOptionsCount         int
	}{
		{
			name:                       "NoAuthNoEncryption",
			config:                     Config{},
			expectAuthUserNameCalls:    []AuthUserNameCall{},
			expectCertificateFileCalls: []CertificateFileCall{},
			expectPrivateKeyFileCalls:  []PrivateKeyFileCall{},
			expectUserTokenType:        ua.UserTokenTypeAnonymous,
			expectMessageSecurityMode:  ua.MessageSecurityModeNone,
			expectPolicy:               "None",
			expectOptionsCount:         1,
		},
		{
			name:                       "AuthNoEncryption",
			config:                     Config{User: "user", Password: "pass"},
			expectAuthUserNameCalls:    []AuthUserNameCall{{user: "user", pass: "pass"}},
			expectCertificateFileCalls: []CertificateFileCall{},
			expectPrivateKeyFileCalls:  []PrivateKeyFileCall{},
			expectUserTokenType:        ua.UserTokenTypeUserName,
			expectMessageSecurityMode:  ua.MessageSecurityModeNone,
			expectPolicy:               "None",
			expectOptionsCount:         2,
		},
		{
			name:                       "NoAuthEncryption",
			config:                     Config{CertFile: "certf.ile", KeyFile: "keyf.ile"},
			expectAuthUserNameCalls:    []AuthUserNameCall{},
			expectCertificateFileCalls: []CertificateFileCall{{filename: "certf.ile"}},
			expectPrivateKeyFileCalls:  []PrivateKeyFileCall{{filename: "keyf.ile"}},
			expectUserTokenType:        ua.UserTokenTypeAnonymous,
			expectMessageSecurityMode:  ua.MessageSecurityModeSignAndEncrypt,
			expectPolicy:               "Basic256Sha256",
			expectOptionsCount:         3,
		},
		{
			name:                       "AuthEncryption",
			config:                     Config{User: "user", Password: "pass", CertFile: "certf.ile", KeyFile: "keyf.ile"},
			expectAuthUserNameCalls:    []AuthUserNameCall{{user: "user", pass: "pass"}},
			expectCertificateFileCalls: []CertificateFileCall{{filename: "certf.ile"}},
			expectPrivateKeyFileCalls:  []PrivateKeyFileCall{{filename: "keyf.ile"}},
			expectUserTokenType:        ua.UserTokenTypeUserName,
			expectMessageSecurityMode:  ua.MessageSecurityModeSignAndEncrypt,
			expectPolicy:               "Basic256Sha256",
			expectOptionsCount:         4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authUsernameCalls := []AuthUserNameCall{}
			certificateFileCalls := []CertificateFileCall{}
			privateKeyFileCalls := []PrivateKeyFileCall{}
			ep := &ua.EndpointDescription{}
			mockedSecurityExtDeps := &SecurityExtDepsMock{
				AuthUsernameFunc: func(user, pass string) opcua.Option {
					return func(c *opcua.Config) {}
				},
				CertificateFileFunc: func(filename string) opcua.Option {
					return func(c *opcua.Config) {}
				},
				PrivateKeyFileFunc: func(filename string) opcua.Option {
					return func(c *opcua.Config) {}
				},
				SecurityFromEndpointFunc: func(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
					return func(c *opcua.Config) {}
				},
			}

			s, err := NewSecurity(&tc.config, mockedSecurityExtDeps)

			if msg := testutils.AssertError(t, err, false); msg != "" {
				t.Fatalf("NewSecurity(): %s", msg)
			}
			// MessageSecurityMode method assertions
			if got, want := s.MessageSecurityMode(), tc.expectMessageSecurityMode; got != want {
				t.Errorf("MessageSecurityMode(): want %v, got %v", want, got)
			}
			// Policy method assertions
			if got, want := s.Policy(), tc.expectPolicy; got != want {
				t.Errorf("Policy(): want %q, got %q", want, got)
			}
			// Options method assertions
			if got, want := len(s.Options(ep)), tc.expectOptionsCount; got != want {
				t.Errorf("Options() count: want %d, got %d", want, got)
			}
			// AuthUserName dependency assertions
			for _, c := range mockedSecurityExtDeps.AuthUsernameCalls() {
				authUsernameCalls = append(authUsernameCalls, AuthUserNameCall{user: c.User, pass: c.Pass})
			}
			if got, want := authUsernameCalls, tc.expectAuthUserNameCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("AuthUsername calls: want %#v, got %#v", want, got)
			}
			// CertificateFile dependency assertions
			for _, c := range mockedSecurityExtDeps.CertificateFileCalls() {
				certificateFileCalls = append(certificateFileCalls, CertificateFileCall{filename: c.Filename})
			}
			if got, want := certificateFileCalls, tc.expectCertificateFileCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("CertificateFile calls: want %#v, got %#v", want, got)
			}
			// PrivateKeyFile dependency assertions
			for _, c := range mockedSecurityExtDeps.PrivateKeyFileCalls() {
				privateKeyFileCalls = append(privateKeyFileCalls, PrivateKeyFileCall{filename: c.Filename})
			}
			if got, want := privateKeyFileCalls, tc.expectPrivateKeyFileCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("PrivateKeyFile calls: want %#v, got %#v", want, got)
			}
			// SecurityFromEndpoint dependency assertions
			if got, want := len(mockedSecurityExtDeps.SecurityFromEndpointCalls()), 1; got != want {
				t.Errorf("SecurityFromEndpoint call count: want %d, got %d", want, got)
			}
			if got, want := mockedSecurityExtDeps.SecurityFromEndpointCalls()[0].Ep, ep; got != want {
				t.Errorf("SecurityFromEndpoint ep argument: want %v, got %v", want, got)
			}
			if got, want := mockedSecurityExtDeps.SecurityFromEndpointCalls()[0].AuthType, tc.expectUserTokenType; got != want {
				t.Errorf("SecurityFromEndpoint authType argument: want %v, got %v", want, got)
			}
		})
	}
}
