package opcua_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func TestMonitorConnect(t *testing.T) {
	cases := []struct {
		name           string
		connectError   bool
		createSubError bool
		expectError    bool
	}{
		{
			name:           "ConnectError",
			connectError:   true,
			createSubError: false,
			expectError:    true,
		},
		{
			name:           "CreateSubscriptionError",
			connectError:   false,
			createSubError: true,
			expectError:    true,
		},
		{
			name:           "Success",
			connectError:   false,
			createSubError: false,
			expectError:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedClientProvider := &ClientProviderMock{
				ConnectFunc: func(ctx context.Context) error {
					if tc.connectError {
						return testutils.ErrTesting
					}
					return nil
				},
			}
			mockedSubscriptionManagerProvider := &SubscriptionManagerProviderMock{
				CreateFunc: func(ctx context.Context, i time.Duration, nc chan<- *opcua.PublishNotificationData) (*Subscription, error) {
					if tc.createSubError {
						return nil, testutils.ErrTesting
					}
					return &Subscription{}, nil
				},
			}
			m := NewMonitor(mockedClientProvider, mockedSubscriptionManagerProvider, nil)

			err := m.Connect(context.Background())

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Error(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := mockedSubscriptionManagerProvider.CreateCalls()[0].I, time.Minute; got != want {
				t.Errorf("Dummy subscription interval: want %v, got %v", want, got)
			}
			subChan := mockedSubscriptionManagerProvider.CreateCalls()[0].Nc
			for i := 1; i <= 50; i++ {
				select {
				case <-time.After(100 * time.Millisecond):
					t.Errorf("Dummy subscription feeding blocked at iteration #%d", i)
					i = 50 // Stop the loop
				case subChan <- &opcua.PublishNotificationData{}:
				}
			}
		})
	}
}

func TestMonitorRead(t *testing.T) {
	cases := []struct {
		name                string
		namespaceIndexError bool
		clientReadError     bool
		badValueStatus      bool
		expectError         bool
		expectValues        *ReadValues
	}{
		{
			name:                "NamespaceIndexError",
			namespaceIndexError: true,
			clientReadError:     false,
			badValueStatus:      false,
			expectError:         true,
			expectValues:        nil,
		},
		{
			name:                "ClientReadError",
			namespaceIndexError: false,
			clientReadError:     true,
			badValueStatus:      false,
			expectError:         true,
			expectValues:        nil,
		},
		{
			name:                "BadValueStatus",
			namespaceIndexError: false,
			clientReadError:     false,
			badValueStatus:      true,
			expectError:         true,
			expectValues:        nil,
		},
		{
			name:                "Success",
			namespaceIndexError: false,
			clientReadError:     false,
			badValueStatus:      false,
			expectError:         false,
			expectValues: &ReadValues{
				Timestamp: time.UnixMicro(0),
				Values: map[string]*ua.Variant{
					"1":     ua.MustVariant(37.5),
					"2":     ua.MustVariant("val2"),
					"node3": ua.MustVariant(int32(42)),
					"node4": ua.MustVariant(true),
				},
			},
		},
	}

	const nj = `[{"namespaceURI":"ns1","nodes":[1,2]},{"namespaceURI":"ns2","nodes":["node3","node4"]}]`
	var no []NodesObject
	if err := json.Unmarshal([]byte(nj), &no); err != nil {
		t.Fatalf("error marshalling: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedClientProvider := &ClientProviderMock{
				FindNamespaceWithContextFunc: func(ctx context.Context, name string) (uint16, error) {
					if tc.namespaceIndexError {
						return 0, testutils.ErrTesting
					}
					return 42, nil
				},
				ReadWithContextFunc: func(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error) {
					if tc.clientReadError {
						return nil, testutils.ErrTesting
					}
					resp := &ua.ReadResponse{
						ResponseHeader: &ua.ResponseHeader{Timestamp: time.UnixMicro(0)},
						Results: []*ua.DataValue{
							{Value: ua.MustVariant(37.5), Status: ua.StatusOK},
							{Value: ua.MustVariant("val2"), Status: ua.StatusOK},
							{Value: ua.MustVariant(int32(42)), Status: ua.StatusOK},
							{Value: ua.MustVariant(true), Status: ua.StatusOK},
						},
					}
					if tc.badValueStatus {
						resp.Results[1].Status = ua.StatusBad
					}
					return resp, nil
				},
			}
			m := NewMonitor(mockedClientProvider, &SubscriptionManagerProviderMock{}, no)

			fields, err := m.Read(context.Background())

			if !tc.namespaceIndexError {
				if got, want := mockedClientProvider.FindNamespaceWithContextCalls()[0].Name, "ns1"; got != want {
					t.Errorf("NamespaceIndex first call nsURI argument: want %q, got %q", want, got)
				}
				if got, want := mockedClientProvider.FindNamespaceWithContextCalls()[1].Name, "ns2"; got != want {
					t.Errorf("NamespaceIndex second call nsURI argument: want %q, got %q", want, got)
				}
			}
			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if got, want := fields, tc.expectValues; !reflect.DeepEqual(got, want) {
				t.Errorf("fields map: want %#v, got %#v", want, got)
			}
		})
	}
}

func TestMonitorSubscribeError(t *testing.T) {
	cases := []struct {
		name                     string
		namespaceIndexError      bool
		namespaceNotFoundError   bool
		createSubError           bool
		monitorError             bool
		monitoredItemCreateError bool
		expectSubCancelCalls     int
	}{
		{
			name:                     "NamespaceIndexError",
			namespaceIndexError:      true,
			createSubError:           false,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "CreateSubError",
			namespaceIndexError:      false,
			createSubError:           true,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "MonitorError",
			namespaceIndexError:      false,
			createSubError:           false,
			monitorError:             true,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "MonitoredItemCreateError",
			namespaceIndexError:      false,
			createSubError:           false,
			monitorError:             false,
			monitoredItemCreateError: true,
			expectSubCancelCalls:     1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedChannelProvider := &ChannelProviderMock{
				IntervalFunc: func() time.Duration { return 0 },
				StringFunc:   func() string { return "" },
			}
			mockedNodeIDProvider := &NodeIDProviderMock{
				NodeIDFunc: func(ns uint16) *ua.NodeID { return ua.NewNumericNodeID(0, 0) },
			}
			mockedSubscriptionProvider := &SubscriptionProviderMock{
				CancelFunc: func(ctx context.Context) error {
					return nil
				},
				MonitorWithContextFunc: func(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
					if tc.monitorError {
						return nil, testutils.ErrTesting
					}
					resp := &ua.CreateMonitoredItemsResponse{
						Results: []*ua.MonitoredItemCreateResult{
							{StatusCode: ua.StatusOK},
							{StatusCode: ua.StatusOK},
							{StatusCode: ua.StatusOK},
						},
					}
					if tc.monitoredItemCreateError {
						resp.Results[1].StatusCode = ua.StatusBadUnexpectedError
					}
					return resp, nil
				},
			}
			mockedClientProvider := &ClientProviderMock{
				FindNamespaceWithContextFunc: func(ctx context.Context, nsURI string) (uint16, error) {
					if tc.namespaceIndexError {
						return 0, testutils.ErrTesting
					}
					return 0, nil
				},
			}
			mockedSubscriptionManagerProvider := &SubscriptionManagerProviderMock{
				CreateFunc: func(ctx context.Context, i time.Duration, nc chan<- *opcua.PublishNotificationData) (*Subscription, error) {
					if tc.createSubError {
						return nil, testutils.ErrTesting
					}
					return NewSubscription(mockedSubscriptionProvider), nil
				},
			}
			m := NewMonitor(mockedClientProvider, mockedSubscriptionManagerProvider, nil)

			err := m.Subscribe(context.Background(), "", mockedChannelProvider, []NodeIDProvider{mockedNodeIDProvider, mockedNodeIDProvider, mockedNodeIDProvider})

			if got, want := len(mockedSubscriptionProvider.CancelCalls()), tc.expectSubCancelCalls; got != want {
				t.Errorf("Cancel call count: want %d, got %d", want, got)
			}
			if got, want := len(m.Subs()), 0; got != want {
				t.Errorf("subscriptions count: want %d, got %d", want, got)
			}
			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}

func newNodeIDProviderMock(t *testing.T, n *ua.NodeID, expNS uint16) *NodeIDProviderMock {
	t.Helper()
	return &NodeIDProviderMock{
		NodeIDFunc: func(ns uint16) *ua.NodeID {
			if got, want := ns, expNS; got != want {
				t.Errorf("NodeID ns argument: want %d, got %d", want, got)
			}
			return n
		},
	}
}

func TestMonitorSubscribeSuccess(t *testing.T) {
	cases := []struct {
		name                  string
		subName               string
		expectNewSubscription bool
	}{
		{
			name:                  "ExistingSubscription",
			subName:               "sub0",
			expectNewSubscription: false,
		},
		{
			name:                  "NewSubscription",
			subName:               "sub1",
			expectNewSubscription: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sentinelNodes := []struct {
				nodeID *ua.NodeID
				handle uint32
			}{
				{nodeID: ua.NewNumericNodeID(2, 42), handle: 1},
				{nodeID: ua.NewStringNodeID(2, "node1"), handle: 2},
				{nodeID: ua.NewStringNodeID(2, "node2"), handle: 3},
			}

			mockedNodeIDProviders := make([]NodeIDProvider, len(sentinelNodes))
			for i, sn := range sentinelNodes {
				mockedNodeIDProviders[i] = newNodeIDProviderMock(t, sn.nodeID, 2)
			}
			mockedChannelProvider := &ChannelProviderMock{
				IntervalFunc: func() time.Duration { return 5 * time.Minute },
				StringFunc:   func() string { return tc.subName },
			}
			mockedSubscriptionProvider := &SubscriptionProviderMock{
				MonitorWithContextFunc: func(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
					for i, item := range items {
						if got, want := item.ItemToMonitor.NodeID, sentinelNodes[i].nodeID; !reflect.DeepEqual(got, want) {
							t.Errorf("Monitor call, %q NodeID: want %q, got %q", item.ItemToMonitor.NodeID, want, got)
						}
						if got, want := item.RequestedParameters.ClientHandle, sentinelNodes[i].handle; got != want {
							t.Errorf("Monitor call, %q node requested client handle: want %d, got %d", item.ItemToMonitor.NodeID, want, got)
						}
					}
					return &ua.CreateMonitoredItemsResponse{}, nil
				},
			}
			mockedClientProvider := &ClientProviderMock{
				FindNamespaceWithContextFunc: func(ctx context.Context, nsURI string) (uint16, error) {
					return 2, nil
				},
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error) {
					if got, want := params.Interval, 5*time.Minute; got != want {
						t.Errorf("Subscribe Interval argument: want %v, got %v", want, got)
					}
					return &opcua.Subscription{}, nil
				},
			}
			mockedSubscriptionManagerProvider := &SubscriptionManagerProviderMock{
				CreateFunc: func(ctx context.Context, i time.Duration, nc chan<- *opcua.PublishNotificationData) (*Subscription, error) {
					return NewSubscription(mockedSubscriptionProvider), nil
				},
			}
			m := NewMonitor(mockedClientProvider, mockedSubscriptionManagerProvider, nil)
			m.AddSubscription("sub0", &SubscriptionProviderMock{})

			err := m.Subscribe(context.Background(), "", mockedChannelProvider, mockedNodeIDProviders)

			var (
				expectCreateCalled       = 0
				expectMonitorCalled      = 0
				expectSubscriptionsCount = 1
			)
			if tc.expectNewSubscription {
				expectCreateCalled = 1
				expectMonitorCalled = 1
				expectSubscriptionsCount = 2

			}
			if got, want := len(mockedSubscriptionManagerProvider.CreateCalls()), expectCreateCalled; got != want {
				t.Errorf("Create call count: want %d, got %d", want, got)
			}
			if got, want := len(mockedSubscriptionProvider.MonitorWithContextCalls()), expectMonitorCalled; got != want {
				t.Errorf("Monitor call count: want %d, got %d", want, got)
			}
			if got, want := len(m.Subs()), expectSubscriptionsCount; got != want {
				t.Errorf("subscriptions count: want %d, got %d", want, got)
			}
			if msg := testutils.AssertError(t, err, false); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}

func TestMonitorGetDataChange(t *testing.T) {
	cases := []struct {
		name             string
		contextCancelled bool
		publish          *opcua.PublishNotificationData
		missingChannel   bool
		expectError      bool
	}{
		{
			name:             "ContextCancelled",
			contextCancelled: true,
			publish: &opcua.PublishNotificationData{
				SubscriptionID: 5616,
				Value:          &ua.DataChangeNotification{},
			},
			missingChannel: false,
			expectError:    true,
		},
		{
			name:             "NotificationDataError",
			contextCancelled: false,
			publish:          &opcua.PublishNotificationData{Error: testutils.ErrTesting},
			missingChannel:   false,
			expectError:      true,
		},
		{
			name:             "EventNotificationList",
			contextCancelled: false,
			publish:          &opcua.PublishNotificationData{Value: &ua.EventNotificationList{}},
			missingChannel:   false,
			expectError:      true,
		},
		{
			name:             "StatusChangeNotification",
			contextCancelled: false,
			publish:          &opcua.PublishNotificationData{Value: &ua.StatusChangeNotification{}},
			missingChannel:   false,
			expectError:      true,
		},
		{
			name:             "CentrifugoChannelNotFound",
			contextCancelled: false,
			publish: &opcua.PublishNotificationData{
				SubscriptionID: 5616,
				Value:          &ua.DataChangeNotification{},
			},
			missingChannel: true,
			expectError:    true,
		},
		{
			name:             "JSONMarshalError",
			contextCancelled: false,
			publish: &opcua.PublishNotificationData{
				SubscriptionID: 5616,
				Value: &ua.DataChangeNotification{
					MonitoredItems: []*ua.MonitoredItemNotification{
						{
							Value: &ua.DataValue{
								Value: ua.MustVariant(math.NaN()),
							},
						},
					},
				},
			},
			missingChannel: false,
			expectError:    true,
		},
		{
			name:             "Success",
			contextCancelled: false,
			publish: &opcua.PublishNotificationData{
				SubscriptionID: 5616,
				Value: &ua.DataChangeNotification{
					MonitoredItems: []*ua.MonitoredItemNotification{
						{ClientHandle: 1, Value: &ua.DataValue{Value: ua.MustVariant("string")}},
						{ClientHandle: 8, Value: &ua.DataValue{Value: ua.MustVariant(uint8(42))}},
						{ClientHandle: 12, Value: &ua.DataValue{Value: ua.MustVariant(time.UnixMilli(0))}},
						{ClientHandle: 13, Value: &ua.DataValue{Value: ua.MustVariant(37.5)}},
					},
				},
			},
			missingChannel: false,
			expectError:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMonitor(&ClientProviderMock{}, &SubscriptionManagerProviderMock{}, nil)
			if !tc.missingChannel {
				mockedSubscriptionProvider := &SubscriptionProviderMock{
					IDFunc: func() uint32 { return 5616 },
				}
				m.AddSubscription("subforchannel", mockedSubscriptionProvider)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.contextCancelled {
				cancel()
			} else {
				m.PushNotification(tc.publish)
			}
			n, d, err := m.GetDataChange(ctx)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := n, "subforchannel"; got != want {
				t.Errorf("channel name: want %q, got %q", want, got)
			}
			expected := map[string]interface{}{
				"0":  "string",
				"7":  42.0,
				"11": "1970-01-01T00:00:00Z",
				"12": 37.5,
			}
			var g map[string]interface{}
			if err := json.Unmarshal(d, &g); err != nil {
				t.Fatalf("error unmarshalling: %v", err)
			}
			if got, want := g, expected; !reflect.DeepEqual(got, want) {
				t.Errorf("JSON data: want %q, got %q", want, got)
			}
		})
	}
}

func TestMonitorPurge(t *testing.T) {
	cases := []struct {
		name                  string
		channels              []string
		cancelError           bool
		expectCancelCallCount int
		expectRemainingSubs   int
		expectErrorCount      int
	}{
		{
			name:                  "NoSubscriptionRemoved",
			channels:              []string{"sub1", "sub2", "sub3"},
			cancelError:           false,
			expectCancelCallCount: 0,
			expectRemainingSubs:   3,
			expectErrorCount:      0,
		},
		{
			name:                  "OneSubscriptionRemovedNoError",
			channels:              []string{"sub2", "sub3"},
			cancelError:           false,
			expectCancelCallCount: 1,
			expectRemainingSubs:   2,
			expectErrorCount:      0,
		},
		{
			name:                  "TwoSubscriptionsRemovedNoError",
			channels:              []string{"sub2"},
			cancelError:           false,
			expectCancelCallCount: 2,
			expectRemainingSubs:   1,
			expectErrorCount:      0,
		},
		{
			name:                  "OneSubscriptionRemovedWithError",
			channels:              []string{"sub1", "sub2"},
			cancelError:           true,
			expectCancelCallCount: 1,
			expectRemainingSubs:   3,
			expectErrorCount:      1,
		},
		{
			name:                  "TwoSubscriptionsRemovedWithError",
			channels:              []string{"sub2"},
			cancelError:           true,
			expectCancelCallCount: 2,
			expectRemainingSubs:   3,
			expectErrorCount:      2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedSubscriptionProvider := &SubscriptionProviderMock{
				CancelFunc: func(ctx context.Context) error {
					if tc.cancelError {
						return testutils.ErrTesting
					}
					return nil
				},
			}
			m := NewMonitor(&ClientProviderMock{}, &SubscriptionManagerProviderMock{}, nil)
			m.AddSubscription("sub1", mockedSubscriptionProvider)
			m.AddSubscription("sub2", &SubscriptionProviderMock{})
			m.AddSubscription("sub3", mockedSubscriptionProvider)

			errs := m.Purge(context.Background(), tc.channels)

			if got, want := len(mockedSubscriptionProvider.CancelCalls()), tc.expectCancelCallCount; got != want {
				t.Errorf("Cancel calls count: want %d, got %d", want, got)
			}
			if got, want := len(m.Subs()), tc.expectRemainingSubs; got != want {
				t.Errorf("remaining subscriptions count: want %d, got %d", want, got)
			}
			if got, want := len(errs), tc.expectErrorCount; got != want {
				t.Errorf("errors count: want %d, got %d", want, got)
			}
		})
	}
}

func TestMonitorHasSubscriptions(t *testing.T) {
	m := NewMonitor(&ClientProviderMock{}, &SubscriptionManagerProviderMock{}, nil)
	if m.HasSubscriptions() {
		t.Error("expected monitor to not have subscriptions")
	}
	m.AddSubscription("chan1", &SubscriptionProviderMock{})
	if !m.HasSubscriptions() {
		t.Error("expected monitor to have subscriptions")
	}
}

func TestMonitorStop(t *testing.T) {
	mockedClientProvider := &ClientProviderMock{
		CloseWithContextFunc: func(ctx context.Context) error {
			return testutils.ErrTesting
		},
	}

	m := NewMonitor(mockedClientProvider, &SubscriptionManagerProviderMock{}, nil)
	m.InitDummyChannel()
	m.SetDummySub(&SubscriptionProviderMock{
		CancelFunc: func(ctx context.Context) error {
			return testutils.ErrTesting
		},
	})
	var mockedSubscriptions [5]*SubscriptionProviderMock
	for i := range mockedSubscriptions {
		mockedSubscriptionProvider := &SubscriptionProviderMock{
			CancelFunc: func(ctx context.Context) error {
				if len(mockedClientProvider.CloseWithContextCalls()) > 0 {
					t.Errorf("client has been closed before subscription cancel call")
				}
				return testutils.ErrTesting
			},
		}
		mockedSubscriptions[i] = mockedSubscriptionProvider
		m.AddSubscription(fmt.Sprintf("sub%d", i), mockedSubscriptionProvider)
	}

	errs := m.Stop(context.Background())

	closed := false
	select {
	case _, ok := <-m.DummyChannel():
		closed = !ok
	default:
	}
	if !closed {
		t.Errorf("Dummy subscription channel unexpectedly not closed")
	}
	if got, want := len(mockedClientProvider.CloseWithContextCalls()), 1; got != want {
		t.Errorf("client.Close call count: want %d, got %d", want, got)
	}
	for _, v := range mockedSubscriptions {
		if got, want := len(v.CancelCalls()), 1; got != want {
			t.Errorf("Subscription.Unsubscribe call count: want %d, got %d", want, got)
		}
	}
	if got, want := len(errs), 7; got != want {
		t.Errorf("errors count: want %d, got %d", want, got)
	}
}

func TestMonitorHealth(t *testing.T) {
	cases := []struct {
		name            string
		state           opcua.ConnState
		expectedHealthy bool
		expectedMessage string
	}{
		{
			name:            "Unhealthy",
			state:           opcua.Connecting,
			expectedHealthy: false,
			expectedMessage: opcua.Connecting.String(),
		},
		{
			name:            "Healthy",
			state:           opcua.Connected,
			expectedHealthy: true,
			expectedMessage: opcua.Connected.String(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedClientProvider := &ClientProviderMock{
				StateFunc: func() opcua.ConnState {
					return tc.state
				},
			}
			m := NewMonitor(mockedClientProvider, &SubscriptionManagerProviderMock{}, nil)

			h, msg := m.Health(context.Background())

			if got, want := h, tc.expectedHealthy; got != want {
				t.Errorf("healthy status: want %v, got %v", want, got)
			}
			if got, want := msg, tc.expectedMessage; got != want {
				t.Errorf("health message: want %q, got %q", want, got)
			}
		})
	}
}
