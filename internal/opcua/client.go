package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
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
	CallWithContext(ctx context.Context, req *ua.CallMethodRequest) (*ua.CallMethodResult, error)
	Connect(context.Context) (err error)
	NamespaceArrayWithContext(ctx context.Context) ([]string, error)
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

// GetMonitoredItems executes the eponymous method on the provided caller.
//
// See https://reference.opcfoundation.org/Core/docs/Part5/9.1
//
// Upon success, it returns a slice of monitored items client handles.
func (c *Client) GetMonitoredItems(ctx context.Context, subID uint32) ([]uint32, error) {
	req := &ua.CallMethodRequest{
		ObjectID:       ua.NewNumericNodeID(0, id.Server),
		MethodID:       ua.NewNumericNodeID(0, id.Server_GetMonitoredItems),
		InputArguments: []*ua.Variant{ua.MustVariant(subID)},
	}

	res, err := c.CallWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error calling the method: %w", err)
	}
	if res.StatusCode != ua.StatusOK {
		return nil, fmt.Errorf("method call failed: %w", res.StatusCode)
	}

	return res.OutputArguments[1].Value().([]uint32), nil
}

// NamespaceIndex returns the index of the provided namespace URI in the server namespace array.
func (c *Client) NamespaceIndex(ctx context.Context, nsURI string) (uint16, error) {
	nsa, err := c.NamespaceArrayWithContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("error getting namespace array: %w", err)
	}

	nsi := -1
	for i, uri := range nsa {
		if uri == nsURI {
			nsi = i
			break
		}
	}
	if nsi == -1 {
		return 0, fmt.Errorf("namespace URI %q not found", nsURI)
	}

	return uint16(nsi), nil
}
