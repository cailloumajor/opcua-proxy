package opcua

import (
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out security_mocks_test.go . SecurityExtDeps

// SecurityExtDeps represents top-level functions of gopcua library related
// to security options.
type SecurityExtDeps interface {
	AuthUsername(user, pass string) opcua.Option
	CertificateFile(filename string) opcua.Option
	PrivateKeyFile(filename string) opcua.Option
	SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option
}

// Security represents OPC-UA security parameters.
type Security struct {
	ed SecurityExtDeps

	utt    ua.UserTokenType
	msm    ua.MessageSecurityMode
	policy string
	opts   []opcua.Option
}

// NewSecurity returns a Security structure populated from given configuration,
// or an error in case of erroneous configuration.
func NewSecurity(cfg *Config, ed SecurityExtDeps) (*Security, error) {
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
		ed:   ed,
		opts: []opcua.Option{},
	}

	if cfg.User == "" {
		s.utt = ua.UserTokenTypeAnonymous
	} else {
		s.utt = ua.UserTokenTypeUserName
		s.opts = append(s.opts, ed.AuthUsername(cfg.User, cfg.Password))
	}

	if cfg.CertFile == "" {
		s.msm = ua.MessageSecurityModeNone
		s.policy = "None"
	} else {
		s.msm = ua.MessageSecurityModeSignAndEncrypt
		s.policy = "Basic256Sha256"
		s.opts = append(
			s.opts,
			ed.CertificateFile(cfg.CertFile),
			ed.PrivateKeyFile(cfg.KeyFile),
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
	return append(s.opts, s.ed.SecurityFromEndpoint(ep, s.utt))
}
