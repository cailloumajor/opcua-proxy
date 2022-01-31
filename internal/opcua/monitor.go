package opcua

import (
	"context"
	"sync"
	"time"
)

//go:generate moq -out monitor_mocks_test.go . ClientProvider Subscription

// ClientProvider is a consumer contract modelling an OPC-UA client provider.
type ClientProvider interface {
	RawClientProvider
	Close() error
}

// Subscription is a consumer contract modelling an OPC-UA subscription.
type Subscription interface {
	Cancel(ctx context.Context) error
}

// Monitor is an OPC-UA node monitor wrapping an client.
type Monitor struct {
	client ClientProvider

	mu   sync.Mutex
	subs map[time.Duration]Subscription
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(ctx context.Context, cfg *Config, c ClientProvider) *Monitor {
	return &Monitor{
		client: c,
		subs:   make(map[time.Duration]Subscription),
	}
}

// Stop cancels all subscriptions and closes the wrapped client.
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
