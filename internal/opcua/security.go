package opcua

import (
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out security_mocks_test.go . SecurityOptsProvider

// SecurityOptsProvider models an OPC-UA security options provider.
type SecurityOptsProvider interface {
	AuthUsername(user, pass string) opcua.Option
	CertificateFile(filename string) opcua.Option
	PrivateKeyFile(filename string) opcua.Option
	SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option
}

// Security represents OPC-UA security parameters.
type Security struct {
	sop SecurityOptsProvider

	utt    ua.UserTokenType
	msm    ua.MessageSecurityMode
	policy string
	opts   []opcua.Option
}

// NewSecurity returns a Security structure populated from given configuration,
// or an error in case of erroneous configuration.
func NewSecurity(cfg *Config, sop SecurityOptsProvider) (*Security, error) {
	if cfg.User == "" && cfg.Password != "" {
		return nil, fmt.Errorf("missing username")
	}
	if cfg.CertFile == "" && cfg.KeyFile != "" {
		return nil, fmt.Errorf("missing certificate file")
	}
	if cfg.CertFile != "" && cfg.KeyFile == "" {
		return nil, fmt.Errorf("missing private key file")
	}

	s := &Security{
		sop:  sop,
		opts: []opcua.Option{},
	}

	if cfg.User == "" {
		s.utt = ua.UserTokenTypeAnonymous
	} else {
		s.utt = ua.UserTokenTypeUserName
		s.opts = append(s.opts, sop.AuthUsername(cfg.User, cfg.Password))
	}

	if cfg.CertFile == "" {
		s.msm = ua.MessageSecurityModeNone
		s.policy = "None"
	} else {
		s.msm = ua.MessageSecurityModeSignAndEncrypt
		s.policy = "Basic256Sha256"
		s.opts = append(
			s.opts,
			sop.CertificateFile(cfg.CertFile),
			sop.PrivateKeyFile(cfg.KeyFile),
		)
	}

	return s, nil
}

// MessageSecurityMode returns the message security mode.
func (s *Security) MessageSecurityMode() ua.MessageSecurityMode {
	return s.msm
}

// Policy returns the message security policy.
func (s *Security) Policy() string {
	return s.policy
}

// Options returns security related OPC-UA options.
func (s *Security) Options(ep *ua.EndpointDescription) []opcua.Option {
	return append(s.opts, s.sop.SecurityFromEndpoint(ep, s.utt))
}

// DefaultSecurityOptsProvider represents the default SecurityOptsProvider implementation.
type DefaultSecurityOptsProvider struct{}

// AuthUsername implements SecurityOptsProvider.
func (d DefaultSecurityOptsProvider) AuthUsername(user, pass string) opcua.Option {
	return opcua.AuthUsername(user, pass)
}

// CertificateFile implements SecurityOptsProvider.
func (d DefaultSecurityOptsProvider) CertificateFile(filename string) opcua.Option {
	return opcua.CertificateFile(filename)
}

// PrivateKeyFile implements SecurityOptsProvider.
func (d DefaultSecurityOptsProvider) PrivateKeyFile(filename string) opcua.Option {
	return opcua.PrivateKeyFile(filename)
}

// SecurityFromEndpoint implements SecurityOptsProvider.
func (d DefaultSecurityOptsProvider) SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
	return opcua.SecurityFromEndpoint(ep, authType)
}
