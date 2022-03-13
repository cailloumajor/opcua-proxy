package opcua_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func TestMonitorSubscribeError(t *testing.T) {
	cases := []struct {
		name                     string
		namespaceIndexError      bool
		namespaceNotFoundError   bool
		subscribeError           bool
		monitorError             bool
		monitoredItemCreateError bool
		expectSubCancelCalls     int
	}{
		{
			name:                     "NamespaceIndexError",
			namespaceIndexError:      true,
			subscribeError:           false,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "SubscribeError",
			namespaceIndexError:      false,
			subscribeError:           true,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "MonitorError",
			namespaceIndexError:      false,
			subscribeError:           false,
			monitorError:             true,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "MonitoredItemCreateError",
			namespaceIndexError:      false,
			subscribeError:           false,
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
			mockedSubscription := &SubscriptionProviderMock{
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
				NamespaceIndexFunc: func(ctx context.Context, nsURI string) (uint16, error) {
					if tc.namespaceIndexError {
						return 0, testutils.ErrTesting
					}
					return 0, nil
				},
				SubscribeFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (SubscriptionProvider, error) {
					if tc.subscribeError {
						return nil, testutils.ErrTesting
					}
					return mockedSubscription, nil
				},
			}
			m := NewMonitor(mockedClientProvider)

			err := m.Subscribe(context.Background(), "", mockedChannelProvider, []string{"node1", "node2", "node3"})

			if got, want := len(mockedSubscription.CancelCalls()), tc.expectSubCancelCalls; got != want {
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

func TestMonitorSubscribeSuccess(t *testing.T) {
	cases := []struct {
		name                  string
		subName               string
		interval              time.Duration
		expectNewSubscription bool
	}{
		{
			name:                  "ExistingSubscription",
			subName:               "sub0",
			interval:              0,
			expectNewSubscription: false,
		},
		{
			name:                  "NewSubscription",
			subName:               "sub1",
			interval:              1,
			expectNewSubscription: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sentinelNodes := [...]struct {
				nodeID string
				handle uint32
			}{
				{nodeID: "node1", handle: 1},
				{nodeID: "node2", handle: 2},
				{nodeID: "node3", handle: 3},
			}
			mockedChannelProvider := &ChannelProviderMock{
				IntervalFunc: func() time.Duration { return tc.interval },
				StringFunc:   func() string { return tc.subName },
			}
			mockedSubscription := &SubscriptionProviderMock{
				IDFunc: func() uint32 { return 56461 },
				MonitorWithContextFunc: func(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
					for i, item := range items {
						if got, want := item.ItemToMonitor.NodeID.Namespace(), uint16(2); got != want {
							t.Errorf("Monitor call, %q node namespace: want %d, got %d", item.ItemToMonitor.NodeID, want, got)
						}
						if got, want := item.ItemToMonitor.NodeID.StringID(), sentinelNodes[i].nodeID; got != want {
							t.Errorf("Monitor call, %q node string ID: want %q, got %q", item.ItemToMonitor.NodeID, want, got)
						}
						if got, want := item.RequestedParameters.ClientHandle, sentinelNodes[i].handle; got != want {
							t.Errorf("Monitor call, %q node requested client handle: want %d, got %d", item.ItemToMonitor.NodeID, want, got)
						}
					}
					return &ua.CreateMonitoredItemsResponse{}, nil
				},
			}
			mockedClientProvider := &ClientProviderMock{
				NamespaceIndexFunc: func(ctx context.Context, nsURI string) (uint16, error) {
					return 2, nil
				},
				SubscribeFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (SubscriptionProvider, error) {
					if got, want := params.Interval, time.Duration(tc.interval); got != want {
						t.Errorf("Subscribe Interval argument: want %v, got %v", want, got)
					}
					return mockedSubscription, nil
				},
			}
			m := NewMonitor(mockedClientProvider)
			m.AddSubscription("sub0", mockedSubscription)

			var nodes []string
			for _, n := range sentinelNodes {
				nodes = append(nodes, n.nodeID)
			}
			err := m.Subscribe(context.Background(), "", mockedChannelProvider, nodes)

			var (
				expectSubscribeCalled    = 0
				expectMonitorCalled      = 0
				expectSubscriptionsCount = 1
			)
			if tc.expectNewSubscription {
				expectSubscribeCalled = 1
				expectMonitorCalled = 1
				expectSubscriptionsCount = 2

			}
			if got, want := len(mockedClientProvider.SubscribeCalls()), expectSubscribeCalled; got != want {
				t.Errorf("Subscribe call count: want %d, got %d", want, got)
			}
			if got, want := len(mockedSubscription.MonitorWithContextCalls()), expectMonitorCalled; got != want {
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
			m := NewMonitor(&ClientProviderMock{})
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
			mockedSubscription := &SubscriptionProviderMock{
				CancelFunc: func(ctx context.Context) error {
					if tc.cancelError {
						return testutils.ErrTesting
					}
					return nil
				},
			}
			m := NewMonitor(&ClientProviderMock{})
			m.AddSubscription("sub1", mockedSubscription)
			m.AddSubscription("sub2", &SubscriptionProviderMock{})
			m.AddSubscription("sub3", mockedSubscription)

			errs := m.Purge(context.Background(), tc.channels)

			if got, want := len(mockedSubscription.CancelCalls()), tc.expectCancelCallCount; got != want {
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

func TestMonitorStop(t *testing.T) {
	mockedClientProvider := &ClientProviderMock{
		CloseWithContextFunc: func(ctx context.Context) error {
			return testutils.ErrTesting
		},
	}

	m := NewMonitor(mockedClientProvider)
	var mockedSubscriptions [5]*SubscriptionProviderMock
	for i := range mockedSubscriptions {
		mockedSubscription := &SubscriptionProviderMock{
			CancelFunc: func(ctx context.Context) error {
				if len(mockedClientProvider.CloseWithContextCalls()) > 0 {
					t.Errorf("client has been closed before subscription cancel call")
				}
				return testutils.ErrTesting
			},
		}
		mockedSubscriptions[i] = mockedSubscription
		m.AddSubscription(fmt.Sprintf("sub%d", i), mockedSubscription)
	}

	errs := m.Stop(context.Background())

	if got, want := len(mockedClientProvider.CloseWithContextCalls()), 1; got != want {
		t.Errorf("client.Close call count: want %d, got %d", want, got)
	}
	for _, v := range mockedSubscriptions {
		if got, want := len(v.CancelCalls()), 1; got != want {
			t.Errorf("Subscription.Unsubscribe call count: want %d, got %d", want, got)
		}
	}
	if got, want := len(errs), 6; got != want {
		t.Errorf("errors count: want %d, got %d", want, got)
	}
}

func TestState(t *testing.T) {
	mockedClientProvider := &ClientProviderMock{
		StateFunc: func() opcua.ConnState {
			return opcua.ConnState(255)
		},
	}
	m := NewMonitor(mockedClientProvider)

	if got, want := m.State(), opcua.ConnState(255); got != want {
		t.Errorf("State method: want %v, got %v", want, got)
	}
}
