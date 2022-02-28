package opcua

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out monitor_mocks_test.go . ClientProvider Subscription

// QueueSize represents the size of the buffered channel for data change notifications.
const QueueSize = 8

// ClientProvider is a consumer contract modelling an OPC-UA client provider.
type ClientProvider interface {
	Close() error
}

// Subscription is a consumer contract modelling an OPC-UA subscription.
type Subscription interface {
	Cancel(ctx context.Context) error
}

// Monitor is an OPC-UA node monitor wrapping an client.
type Monitor struct {
	client ClientProvider

	notifyCh chan *opcua.PublishNotificationData

	mu    sync.RWMutex
	subs  map[PublishingInterval]Subscription
	items map[uint32]string
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(ctx context.Context, cfg *Config, c ClientProvider) *Monitor {
	return &Monitor{
		client:   c,
		notifyCh: make(chan *opcua.PublishNotificationData, QueueSize),
		subs:     make(map[PublishingInterval]Subscription),
		items:    make(map[uint32]string),
	}
}

// GetDataChange returns a JSON string from the next dequeued data change notification.
func (m *Monitor) GetDataChange() (string, error) {
	notif := <-m.notifyCh

	if notif.Error != nil {
		return "", fmt.Errorf("notification data error: %w", notif.Error)
	}

	d, ok := notif.Value.(*ua.DataChangeNotification)
	if !ok {
		return "", fmt.Errorf("not a data change notification")
	}

	im := make(map[string]interface{})
	m.mu.RLock()
	for _, mi := range d.MonitoredItems {
		n := m.items[mi.ClientHandle]
		im[n] = mi.Value.Value.Value()
	}
	m.mu.RUnlock()

	j, err := json.Marshal(im)
	if err != nil {
		return "", fmt.Errorf("JSON marshalling error: %w", err)
	}

	return string(j), nil
}

// Stop cancels all subscriptions and closes the wrapped client.
//
// Monitor must not be used after calling Stop().
func (m *Monitor) Stop(ctx context.Context) []error {
	var errs []error

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range m.subs {
		if err := v.Cancel(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if err := m.client.Close(); err != nil {
		errs = append(errs, err)
	}

	return errs
}
