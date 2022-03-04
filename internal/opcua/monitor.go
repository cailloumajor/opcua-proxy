package opcua

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out monitor_mocks_test.go . ClientProvider Subscription

// QueueSize represents the size of the buffered channel for data change notifications.
const QueueSize = 8

// ClientProvider is a consumer contract modelling an OPC-UA client provider.
type ClientProvider interface {
	CloseWithContext(ctx context.Context) error
	NamespaceIndex(ctx context.Context, nsURI string) (uint16, error)
	SubscribeWithContext(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error)
}

// Subscription is a consumer contract modelling an OPC-UA subscription.
type Subscription interface {
	Cancel(ctx context.Context) error
	MonitorWithContext(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error)
}

// Monitor is an OPC-UA node monitor wrapping an client.
type Monitor struct {
	client ClientProvider

	notifyCh chan *opcua.PublishNotificationData

	mu    sync.RWMutex
	subs  map[time.Duration]Subscription
	items map[uint32]string
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(cfg *Config, c ClientProvider) *Monitor {
	return &Monitor{
		client:   c,
		notifyCh: make(chan *opcua.PublishNotificationData, QueueSize),
		subs:     make(map[time.Duration]Subscription),
		items:    make(map[uint32]string),
	}
}

// Subscribe subscribes for nodes data changes on the server.
//
// Provided nodes are string node identifiers.
func (m *Monitor) Subscribe(ctx context.Context, interval time.Duration, nsURI string, nodes ...string) error {
	nsi, err := m.client.NamespaceIndex(ctx, nsURI)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sub, ok := m.subs[interval]
	if !ok {
		p := &opcua.SubscriptionParameters{
			Interval: interval,
		}
		sub, err = m.client.SubscribeWithContext(ctx, p, m.notifyCh)
		if err != nil {
			return fmt.Errorf("error creating subscription: %w", err)
		}
	}

	im := make(map[uint32]string)
	for k, v := range m.items {
		im[k] = v
	}

	reqs := make([]*ua.MonitoredItemCreateRequest, len(nodes))
	for i, node := range nodes {
		handle := uint32(len(im))
		reqs[i] = &ua.MonitoredItemCreateRequest{
			ItemToMonitor: &ua.ReadValueID{
				NodeID:      ua.NewStringNodeID(nsi, node),
				AttributeID: ua.AttributeIDValue,
			},
			MonitoringMode: ua.MonitoringModeReporting,
			RequestedParameters: &ua.MonitoringParameters{
				ClientHandle:     handle,
				SamplingInterval: -1,
				QueueSize:        1,
			},
		}
		im[handle] = node
	}

	res, err := sub.MonitorWithContext(ctx, ua.TimestampsToReturnNeither, reqs...)
	if err != nil {
		return fmt.Errorf("error creating monitored items: %w", err)
	}
	for i, s := range res.Results {
		if s.StatusCode != ua.StatusOK {
			_ = sub.Cancel(ctx) // Desperate attempt...
			return fmt.Errorf("error creating %q monitored item: %w", nodes[i], err)
		}
	}

	m.subs[interval] = sub
	m.items = im

	return nil
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

// Purge unsubscribes and removes subscriptions for intervals that do not exist in provided slice.
func (m *Monitor) Purge(ctx context.Context, intervals []time.Duration) (errs []error) {
	is := make(map[time.Duration]bool)
	for _, interval := range intervals {
		is[interval] = true
	}

	for interval, sub := range m.subs {
		if !is[interval] {
			if err := sub.Cancel(ctx); err != nil {
				errs = append(errs, err)
				continue
			}
			delete(m.subs, interval)
		}
	}

	return
}

// Stop cancels all subscriptions and closes the wrapped client.
//
// Monitor must not be used after calling Stop().
func (m *Monitor) Stop(ctx context.Context) (errs []error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range m.subs {
		if err := v.Cancel(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if err := m.client.CloseWithContext(ctx); err != nil {
		errs = append(errs, err)
	}

	return
}
