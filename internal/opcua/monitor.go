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

//go:generate moq -out monitor_mocks_test.go . ClientProvider SubscriptionProvider ChannelProvider NodeIDProvider

// QueueSize represents the size of the buffered channel for data change notifications.
const QueueSize = 8

// ClientProvider is a consumer contract modelling an OPC-UA client provider.
type ClientProvider interface {
	Close(ctx context.Context) (errs []error)
	NamespaceIndex(ctx context.Context, nsURI string) (uint16, error)
	Read(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error)
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
	Interval() time.Duration
	fmt.Stringer
}

// NodeIDProvider is a consumer contract modelling an OPC-UA NodeID provider.
type NodeIDProvider interface {
	NodeID(ns uint16) *ua.NodeID
}

// ReadValues represents the data values of nodes read from OPC-UA server.
type ReadValues struct {
	Timestamp time.Time
	Values    map[string]*ua.Variant
}

// Monitor represents an OPC-UA monitor.
type Monitor struct {
	client    ClientProvider
	readNodes []NodesObject

	notifyCh chan *opcua.PublishNotificationData

	mu   sync.RWMutex
	subs map[string]SubscriptionProvider // map of subscription by Centrifugo channel name
}

// NewMonitor creates an OPC-UA node monitor.
func NewMonitor(c ClientProvider, n []NodesObject) *Monitor {
	return &Monitor{
		client:    c,
		readNodes: n,
		notifyCh:  make(chan *opcua.PublishNotificationData, QueueSize),
		subs:      make(map[string]SubscriptionProvider),
	}
}

// Read reads configured nodes data values and returns a map of fields.
func (m *Monitor) Read(ctx context.Context) (*ReadValues, error) {
	req := &ua.ReadRequest{
		MaxAge:             0,
		TimestampsToReturn: ua.TimestampsToReturnNeither,
	}

	var de ua.QualifiedName
	var nid []string

	for _, no := range m.readNodes {
		nsi, err := m.client.NamespaceIndex(ctx, no.NamespaceURI)
		if err != nil {
			return nil, err
		}

		for _, n := range no.Nodes {
			nid = append(nid, n.String())
			ntr := &ua.ReadValueID{
				NodeID:       n.NodeID(nsi),
				AttributeID:  ua.AttributeIDValue,
				DataEncoding: &de,
			}
			req.NodesToRead = append(req.NodesToRead, ntr)
		}
	}

	resp, err := m.client.Read(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error reading OPC-UA nodes: %w", err)
	}

	rv := &ReadValues{
		Timestamp: resp.ResponseHeader.Timestamp,
		Values:    make(map[string]*ua.Variant),
	}

	for i, dv := range resp.Results {
		if dv.Status != ua.StatusOK {
			return nil, fmt.Errorf("status error for node %q: %w", req.NodesToRead[i].NodeID.String(), dv.Status)
		}
		rv.Values[nid[i]] = dv.Value
	}

	return rv, nil
}

// Subscribe subscribes for nodes data changes on the server.
func (m *Monitor) Subscribe(ctx context.Context, nsURI string, ch ChannelProvider, nodes []NodeIDProvider) error {
	nsi, err := m.client.NamespaceIndex(ctx, nsURI)
	if err != nil {
		return err
	}

	m.mu.RLock()
	_, exists := m.subs[ch.String()]
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

	reqs := make([]*ua.MonitoredItemCreateRequest, len(nodes))
	for i, node := range nodes {
		reqs[i] = &ua.MonitoredItemCreateRequest{
			ItemToMonitor: &ua.ReadValueID{
				NodeID:       node.NodeID(nsi),
				AttributeID:  ua.AttributeIDValue,
				DataEncoding: &ua.QualifiedName{},
			},
			MonitoringMode: ua.MonitoringModeReporting,
			RequestedParameters: &ua.MonitoringParameters{
				ClientHandle:     uint32(i + 1),
				SamplingInterval: -1,
				QueueSize:        1,
			},
		}
	}

	res, err := sub.MonitorWithContext(ctx, ua.TimestampsToReturnNeither, reqs...)
	if err != nil {
		return fmt.Errorf("error creating monitored items: %w", err)
	}
	for i, s := range res.Results {
		if s.StatusCode != ua.StatusOK {
			_ = sub.Cancel(ctx) // Desperate attempt...
			return fmt.Errorf("error creating %q monitored item: %w", reqs[i].ItemToMonitor.NodeID, s.StatusCode)
		}
	}

	m.subs[ch.String()] = sub

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

	m.mu.RLock()
	const notFoundChannel = ":"
	channel := notFoundChannel
	for ch, sub := range m.subs {
		if notif.SubscriptionID == sub.ID() {
			channel = ch
			break
		}
	}
	m.mu.RUnlock()
	if channel == notFoundChannel {
		return "", nil, fmt.Errorf("Centrifugo channel not found for subscription with ID %d", notif.SubscriptionID)
	}

	im := make(map[uint32]interface{})
	for _, mi := range d.MonitoredItems {
		im[mi.ClientHandle-1] = mi.Value.Value.Value()
	}

	j, err := json.Marshal(im)
	if err != nil {
		return "", nil, fmt.Errorf("JSON marshalling error: %w", err)
	}

	return channel, j, nil
}

// Purge unsubscribes and removes subscriptions for Centrifugo channels that do not exist in provided slice.
func (m *Monitor) Purge(ctx context.Context, channels []string) (errs []error) {
	is := make(map[string]bool)
	for _, ch := range channels {
		is[ch] = true
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for ch, sub := range m.subs {
		if !is[ch] {
			if err := sub.Cancel(ctx); err != nil {
				errs = append(errs, err)
				continue
			}
			delete(m.subs, ch)
		}
	}

	return
}

// HasSubscriptions returns whether the monitor has subscriptions or not.
func (m *Monitor) HasSubscriptions() bool {
	return len(m.subs) > 0
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

	closeErrs := m.client.Close(ctx)
	errs = append(errs, closeErrs...)

	return
}

// State returns the wrapped client connection state.
func (m *Monitor) State() opcua.ConnState {
	return m.client.State()
}
