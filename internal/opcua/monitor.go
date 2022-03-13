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

//go:generate moq -out monitor_mocks_test.go . ClientProvider SubscriptionProvider ChannelProvider

// QueueSize represents the size of the buffered channel for data change notifications.
const QueueSize = 8

// ClientProvider is a consumer contract modelling an OPC-UA client provider.
type ClientProvider interface {
	CloseWithContext(ctx context.Context) error
	NamespaceIndex(ctx context.Context, nsURI string) (uint16, error)
	Subscribe(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (SubscriptionProvider, error)
	State() opcua.ConnState
}

// SubscriptionProvider is a consumer contract modelling an OPC-UA subscription.
type SubscriptionProvider interface {
	Cancel(ctx context.Context) error
	ID() uint32
	MonitorWithContext(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error)
}

// ChannelProvider is a consumer contract modelling a Centrifugo channel.
type ChannelProvider interface {
	Name() string
	Interval() time.Duration
	fmt.Stringer
}

// subShape models the characteristics of a subscription.
type subShape struct {
	name     string
	interval time.Duration
}

// Monitor is an OPC-UA node monitor wrapping a client.
type Monitor struct {
	client ClientProvider

	notifyCh chan *opcua.PublishNotificationData

	mu    sync.RWMutex
	subs  map[subShape]SubscriptionProvider
	chans map[uint32]ChannelProvider
	items map[uint32]string
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(c ClientProvider) *Monitor {
	return &Monitor{
		client:   c,
		notifyCh: make(chan *opcua.PublishNotificationData, QueueSize),
		subs:     make(map[subShape]SubscriptionProvider),
		chans:    make(map[uint32]ChannelProvider),
		items:    make(map[uint32]string),
	}
}

// Subscribe subscribes for nodes data changes on the server.
//
// Provided nodes are string node identifiers.
func (m *Monitor) Subscribe(ctx context.Context, nsURI string, ch ChannelProvider, nodes []string) error {
	nsi, err := m.client.NamespaceIndex(ctx, nsURI)
	if err != nil {
		return err
	}

	si := subShape{
		name:     ch.Name(),
		interval: ch.Interval(),
	}

	m.mu.RLock()
	_, exists := m.subs[si]
	m.mu.RUnlock()

	if exists {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	p := &opcua.SubscriptionParameters{
		Interval: ch.Interval(),
	}
	sub, err := m.client.Subscribe(ctx, p, m.notifyCh)
	if err != nil {
		return fmt.Errorf("error creating subscription: %w", err)
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
				NodeID:       ua.NewStringNodeID(nsi, node),
				AttributeID:  ua.AttributeIDValue,
				DataEncoding: &ua.QualifiedName{},
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
			return fmt.Errorf("error creating %q monitored item: %w", nodes[i], s.StatusCode)
		}
	}

	m.subs[si] = sub
	m.items = im
	m.chans[sub.ID()] = ch

	return nil
}

// GetDataChange returns the Centrifugo channel name and JSON data from
// the next dequeued data change notification.
func (m *Monitor) GetDataChange(ctx context.Context) (string, []byte, error) {
	var notif *opcua.PublishNotificationData
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	case notif = <-m.notifyCh:
	}

	if notif.Error != nil {
		return "", nil, fmt.Errorf("notification data error: %w", notif.Error)
	}

	d, ok := notif.Value.(*ua.DataChangeNotification)
	if !ok {
		return "", nil, fmt.Errorf("not a data change notification")
	}

	im := make(map[string]interface{})
	m.mu.RLock()

	c, ok := m.chans[notif.SubscriptionID]
	if !ok {
		return "", nil, fmt.Errorf("Centrifugo channel not found")
	}

	for _, mi := range d.MonitoredItems {
		n := m.items[mi.ClientHandle]
		im[n] = mi.Value.Value.Value()
	}

	m.mu.RUnlock()

	j, err := json.Marshal(im)
	if err != nil {
		return "", nil, fmt.Errorf("JSON marshalling error: %w", err)
	}

	return c.Name(), j, nil
}

// Purge unsubscribes and removes subscriptions for intervals that do not exist in provided slice.
func (m *Monitor) Purge(ctx context.Context, intervals []time.Duration) (errs []error) {
	is := make(map[time.Duration]bool)
	for _, interval := range intervals {
		is[interval] = true
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, sub := range m.subs {
		if !is[id.interval] {
			if err := sub.Cancel(ctx); err != nil {
				errs = append(errs, err)
				continue
			}
			delete(m.subs, id)
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

// State returns the wrapped client connection state.
func (m *Monitor) State() opcua.ConnState {
	return m.client.State()
}
