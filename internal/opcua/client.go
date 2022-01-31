package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out client_mocks_test.go . ClientExtDeps RawClientProvider SecurityProvider

// ClientExtDeps is a consumer contract modelling external dependencies.
type ClientExtDeps interface {
	GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error)
	SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription
	NewClient(endpoint string, opts ...opcua.Option) RawClientProvider
}

// RawClientProvider is a consumer contract modelling a raw OPC-UA client.
type RawClientProvider interface {
	Connect(context.Context) (err error)
}

// SecurityProvider is a consumer contract modelling an OPC-UA security provider.
type SecurityProvider interface {
	MessageSecurityMode() ua.MessageSecurityMode
	Policy() string
	Options(ep *ua.EndpointDescription) []opcua.Option
}

// Client represents an OPC-UA client connected to a server.
type Client struct {
	RawClientProvider
}

// NewClient creates a new client and connects it to a server.
func NewClient(ctx context.Context, cfg *Config, deps ClientExtDeps, sec SecurityProvider) (*Client, error) {
	eps, err := deps.GetEndpoints(ctx, cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("error getting endpoints: %w", err)
	}

	ep := deps.SelectEndpoint(eps, sec.Policy(), sec.MessageSecurityMode())
	if ep == nil {
		return nil, fmt.Errorf("failed to select an endpoint")
	}

	opts := sec.Options(ep)
	c := deps.NewClient(ep.EndpointURL, opts...)

	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		RawClientProvider: c,
	}, nil
}
