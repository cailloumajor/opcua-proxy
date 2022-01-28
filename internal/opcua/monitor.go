package opcua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out monitor_mocks_test.go . Client Subscription MonitorExtDeps SecurityProvider

// Client is a consumer contract modelling an OPC-UA client.
type Client interface {
	Connect(context.Context) (err error)
	Close() error
}

// Subscription is a consumer contract modelling an OPC-UA subscription.
type Subscription interface {
	Cancel(ctx context.Context) error
}

// MonitorExtDeps is a consumer contract modelling external dependencies.
type MonitorExtDeps interface {
	GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error)
	SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription
	NewClient(endpoint string, opts ...opcua.Option) Client
}

// SecurityProvider is a consumer contract modelling an OPC-UA security provider.
type SecurityProvider interface {
	MessageSecurityMode() ua.MessageSecurityMode
	Policy() string
	Options(ep *ua.EndpointDescription) []opcua.Option
}

// Config holds the OPC-UA part of the configuration.
type Config struct {
	ServerURL string
	User      string
	Password  string
	CertFile  string
	KeyFile   string
}

// Monitor is an OPC-UA node monitor.
type Monitor struct {
	client Client

	mu   sync.Mutex
	subs map[time.Duration]Subscription
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(ctx context.Context, cfg *Config, deps MonitorExtDeps, sec SecurityProvider) (*Monitor, error) {
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

	return &Monitor{
		client: c,
		subs:   make(map[time.Duration]Subscription),
	}, nil
}

// Stop does all the needed job to stop OPC-UA related elements.
func (m *Monitor) Stop(ctx context.Context) []error {
	var errs []error

	if err := m.client.Close(); err != nil {
		errs = append(errs, err)
	}

	for _, v := range m.subs {
		if err := v.Cancel(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs = make(map[time.Duration]Subscription)

	return errs
}
