package opcua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out opcua_mocks_test.go . Client NodeMonitor Subscription NewMonitorDeps SecurityProvider

// Client models an OPC-UA client.
type Client interface {
	Connect(context.Context) (err error)
	Close() error
}

// NodeMonitor models an OPC-UA node monitor.
type NodeMonitor interface {
	SetErrorHandler(cb monitor.ErrHandler)
}

// Subscription models an OPC-UA subscription.
type Subscription interface {
	Unsubscribe(ctx context.Context) error
}

// NewMonitorDeps models the dependencies of NewMonitor.
type NewMonitorDeps interface {
	GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error)
	SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription
	NewClient(endpoint string, opts ...opcua.Option) Client
	NewNodeMonitor(client Client) (NodeMonitor, error)
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
	client      Client
	nodeMonitor NodeMonitor

	mu   sync.Mutex
	subs map[time.Duration]Subscription
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(ctx context.Context, cfg *Config, deps NewMonitorDeps, sec SecurityProvider, logger log.Logger) (*Monitor, error) {
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

	nm, err := deps.NewNodeMonitor(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create a node monitor: %w", err)
	}

	nm.SetErrorHandler(func(c *opcua.Client, s *monitor.Subscription, e error) {
		level.Info(logger).Log("from", "subscription", "sub_id", s.SubscriptionID(), "err", e)
	})

	return &Monitor{
		client:      c,
		nodeMonitor: nm,
		subs:        make(map[time.Duration]Subscription),
	}, nil
}

// Stop does all the needed job to stop OPC-UA related elements.
func (m *Monitor) Stop(ctx context.Context) []error {
	var errs []error

	if err := m.client.Close(); err != nil {
		errs = append(errs, err)
	}

	for _, v := range m.subs {
		if err := v.Unsubscribe(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs = make(map[time.Duration]Subscription)

	return errs
}
