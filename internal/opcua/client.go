package opcua

import (
	"context"
	"fmt"
	"time"

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
	CloseWithContext(ctx context.Context) error
	Connect(context.Context) (err error)
	NamespaceArrayWithContext(ctx context.Context) ([]string, error)
	State() opcua.ConnState
	SubscribeWithContext(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error)
}

// SecurityProvider is a consumer contract modelling an OPC-UA security provider.
type SecurityProvider interface {
	MessageSecurityMode() ua.MessageSecurityMode
	Policy() string
	Options(ep *ua.EndpointDescription) []opcua.Option
}

// Client represents an OPC-UA client connected to a server.
type Client struct {
	inner    RawClientProvider
	dummySub *opcua.Subscription
	stopChan chan struct{}
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

	nc := make(chan *opcua.PublishNotificationData)

	sp := &opcua.SubscriptionParameters{
		Interval: time.Minute,
	}
	ds, err := c.SubscribeWithContext(ctx, sp, nc)
	if err != nil {
		return nil, fmt.Errorf("error creating dummy subscription: %w", err)
	}

	sc := make(chan struct{})

	go func(notifyCh <-chan *opcua.PublishNotificationData, stopChan <-chan struct{}) {
		for {
			select {
			case <-stopChan:
				return
			case <-notifyCh: // drop notifications
			}
		}
	}(nc, sc)

	return &Client{
		inner:    c,
		dummySub: ds,
		stopChan: sc,
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

	res, err := c.inner.CallWithContext(ctx, req)
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
	nsa, err := c.inner.NamespaceArrayWithContext(ctx)
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

// Close wraps inner client close.
func (c *Client) Close(ctx context.Context) (errs []error) {
	if err := c.dummySub.Cancel(ctx); err != nil {
		errs = append(errs, err)
	}

	c.stopChan <- struct{}{}

	if err := c.inner.CloseWithContext(ctx); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// State stub.
func (c *Client) State() opcua.ConnState {
	return c.inner.State()
}

// Subscribe stub.
func (c *Client) Subscribe(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (SubscriptionProvider, error) {
	s, err := c.inner.SubscribeWithContext(ctx, params, notifyCh)
	if err != nil {
		return nil, err
	}
	return &Subscription{inner: s}, nil
}

// DefaultClientExtDeps represents the default ClientExtDeps implementation.
type DefaultClientExtDeps struct{}

// GetEndpoints implements ClientExtDeps.
func (DefaultClientExtDeps) GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
	return opcua.GetEndpoints(ctx, endpoint, opts...)
}

// SelectEndpoint implements ClientExtDeps.
func (DefaultClientExtDeps) SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
	return opcua.SelectEndpoint(endpoints, policy, mode)
}

// NewClient implements ClientExtDeps.
func (DefaultClientExtDeps) NewClient(endpoint string, opts ...opcua.Option) RawClientProvider {
	return opcua.NewClient(endpoint, opts...)
}
