package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out client_mocks_test.go . ClientExtDeps

// ClientExtDeps represents top-level functions of gopcua library related to OPC-UA client.
type ClientExtDeps interface {
	GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error)
	SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription
	NewClient(endpoint string, opts ...opcua.Option) *opcua.Client
	AuthUsername(user, pass string) opcua.Option
	CertificateFile(filename string) opcua.Option
	PrivateKeyFile(filename string) opcua.Option
	SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option
}

// Client wraps an OPC-UA client.
type Client struct {
	*opcua.Client
}

// NewClient creates a configured OPC-UA client.
func NewClient(ctx context.Context, cfg *Config, deps ClientExtDeps) (*Client, error) {
	if deps == nil {
		deps = &defaultClientExtDeps{}
	}

	eps, err := deps.GetEndpoints(ctx, cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("error getting endpoints: %w", err)
	}

	var (
		utt  ua.UserTokenType
		opts []opcua.Option
		msm  ua.MessageSecurityMode
		pol  string
	)

	if cfg.User == "" && cfg.Password == "" {
		utt = ua.UserTokenTypeAnonymous
	} else {
		utt = ua.UserTokenTypeUserName
		opts = append(opts, deps.AuthUsername(cfg.User, cfg.Password))
	}

	if cfg.CertFile == "" && cfg.KeyFile == "" {
		msm = ua.MessageSecurityModeNone
		pol = "None"
	} else {
		msm = ua.MessageSecurityModeSignAndEncrypt
		pol = "Basic256Sha256"
		opts = append(
			opts,
			deps.CertificateFile(cfg.CertFile),
			deps.PrivateKeyFile(cfg.KeyFile),
		)
	}

	ep := deps.SelectEndpoint(eps, pol, msm)
	if ep == nil {
		return nil, fmt.Errorf("failed to select an endpoint")
	}

	opts = append(opts, deps.SecurityFromEndpoint(ep, utt))

	c := deps.NewClient(ep.EndpointURL, opts...)

	return &Client{c}, nil
}
