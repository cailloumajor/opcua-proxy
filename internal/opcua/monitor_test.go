package opcua_test

import (
	"context"
	"math"
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
			mockedSubscription := &SubscriptionMock{
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
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error) {
					if tc.subscribeError {
						return nil, testutils.ErrTesting
					}
					return mockedSubscription, nil
				},
			}
			m := NewMonitor(&Config{}, mockedClientProvider)

			err := m.Subscribe(context.Background(), 0, "", "node1", "node2", "node3")

			if got, want := len(mockedSubscription.CancelCalls()), tc.expectSubCancelCalls; got != want {
				t.Errorf("Cancel call count: want %d, got %d", want, got)
			}
			if got, want := len(m.Subs()), 0; got != want {
				t.Errorf("subscriptions count: want %d, got %d", want, got)
			}
			if got, want := len(m.Items()), 0; got != want {
				t.Errorf("monitored items count: want %d, got %d", want, got)
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
		interval              time.Duration
		expectSubscribeCalled bool
	}{
		{
			name:                  "SubscribeNotCalled",
			interval:              0,
			expectSubscribeCalled: false,
		},
		{
			name:                  "SubscribeCalled",
			interval:              1,
			expectSubscribeCalled: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedSubscription := &SubscriptionMock{
				MonitorWithContextFunc: func(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
					return &ua.CreateMonitoredItemsResponse{}, nil
				},
			}
			mockedClientProvider := &ClientProviderMock{
				NamespaceIndexFunc: func(ctx context.Context, nsURI string) (uint16, error) {
					return 2, nil
				},
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error) {
					return mockedSubscription, nil
				},
			}
			m := NewMonitor(&Config{}, mockedClientProvider)
			m.AddSubscription(0, mockedSubscription)
			m.AddMonitoredItems("existing1", "existing2")
			nodes := []string{"node1", "node2", "node3"}

			err := m.Subscribe(context.Background(), tc.interval, "", nodes...)

			if tc.expectSubscribeCalled {
				if got, want := len(mockedClientProvider.SubscribeWithContextCalls()), 1; got != want {
					t.Errorf("Subscribe call count: want %d, got %d", want, got)
				}
				if got, want := mockedClientProvider.SubscribeWithContextCalls()[0].Params.Interval, time.Duration(tc.interval); got != want {
					t.Errorf("Subscribe Interval argument: want %v, got %v", want, got)
				}
				if got, want := mockedClientProvider.SubscribeWithContextCalls()[0].NotifyCh, m.NotifyChannel(); got != want {
					t.Errorf("Subscribe NotifyCh argument: want %#v, got %#v", want, got)
				}
			} else {
				if got, want := len(mockedClientProvider.SubscribeWithContextCalls()), 0; got != want {
					t.Errorf("Subscribe call count: want %d, got %d", want, got)
				}
			}
			if got, want := len(mockedSubscription.MonitorWithContextCalls()), 1; got != want {
				t.Errorf("Monitor call count: want %d, got %d", want, got)
			}
			for i, item := range mockedSubscription.MonitorWithContextCalls()[0].Items {
				if got, want := item.ItemToMonitor.NodeID.Namespace(), uint16(2); got != want {
					t.Errorf("Monitor call, %q node namespace: want %d, got %d", nodes[i], want, got)
				}
				if got, want := item.ItemToMonitor.NodeID.StringID(), nodes[i]; got != want {
					t.Errorf("Monitor call, %q node string ID: want %q, got %q", nodes[i], want, got)
				}
				if got, want := item.RequestedParameters.ClientHandle, uint32(i+2); got != want {
					t.Errorf("Monitor call, %q node requested client handle: want %d, got %d", nodes[i], want, got)
				}
			}
			expectedSubsCount := 1
			if tc.expectSubscribeCalled {
				expectedSubsCount = 2
			}
			if got, want := len(m.Subs()), expectedSubsCount; got != want {
				t.Errorf("subscriptions count: want %d, got %d", want, got)
			}
			if got, want := len(m.Items()), 5; got != want {
				t.Errorf("monitored items count: want %d, got %d", want, got)
			}
			if msg := testutils.AssertError(t, err, false); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}

func TestMonitorGetDataChange(t *testing.T) {
	cases := []struct {
		name        string
		publish     *opcua.PublishNotificationData
		expectError bool
		expectJSON  string
	}{
		{
			name:        "NotificationDataError",
			publish:     &opcua.PublishNotificationData{Error: testutils.ErrTesting},
			expectError: true,
			expectJSON:  "",
		},
		{
			name:        "EventNotificationList",
			publish:     &opcua.PublishNotificationData{Value: &ua.EventNotificationList{}},
			expectError: true,
			expectJSON:  "",
		},
		{
			name:        "StatusChangeNotification",
			publish:     &opcua.PublishNotificationData{Value: &ua.StatusChangeNotification{}},
			expectError: true,
			expectJSON:  "",
		},
		{
			name: "JSONMarshalError",
			publish: &opcua.PublishNotificationData{
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
			expectError: true,
			expectJSON:  "",
		},
		{
			name: "Success",
			publish: &opcua.PublishNotificationData{
				Value: &ua.DataChangeNotification{
					MonitoredItems: []*ua.MonitoredItemNotification{
						{ClientHandle: 0, Value: &ua.DataValue{Value: ua.MustVariant("string")}},
						{ClientHandle: 1, Value: &ua.DataValue{Value: ua.MustVariant(uint8(42))}},
						{ClientHandle: 2, Value: &ua.DataValue{Value: ua.MustVariant(time.UnixMilli(0))}},
						{ClientHandle: 3, Value: &ua.DataValue{Value: ua.MustVariant(37.5)}},
					},
				},
			},
			expectError: false,
			expectJSON:  `{"node0":"string","node1":42,"node2":"1970-01-01T00:00:00Z","node3":37.5}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMonitor(&Config{}, &ClientProviderMock{})
			m.AddMonitoredItems("node0", "node1", "node2", "node3")
			m.PushNotification(tc.publish)

			d, err := m.GetDataChange()

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if got, want := d, tc.expectJSON; got != want {
				t.Errorf("JSON data: want %q, got %q", want, got)
			}
		})
	}
}

func TestMonitorPurge(t *testing.T) {
	cases := []struct {
		name                  string
		intervals             []time.Duration
		cancelError           bool
		expectCancelCallCount int
		expectRemainingSubs   int
		expectErrorCount      int
	}{
		{
			name:                  "NoSubscriptionRemoved",
			intervals:             []time.Duration{1, 2, 3},
			cancelError:           false,
			expectCancelCallCount: 0,
			expectRemainingSubs:   3,
			expectErrorCount:      0,
		},
		{
			name:                  "OneSubscriptionRemovedNoError",
			intervals:             []time.Duration{2, 3},
			cancelError:           false,
			expectCancelCallCount: 1,
			expectRemainingSubs:   2,
			expectErrorCount:      0,
		},
		{
			name:                  "TwoSubscriptionsRemovedNoError",
			intervals:             []time.Duration{2},
			cancelError:           false,
			expectCancelCallCount: 2,
			expectRemainingSubs:   1,
			expectErrorCount:      0,
		},
		{
			name:                  "OneSubscriptionRemovedWithError",
			intervals:             []time.Duration{1, 2},
			cancelError:           true,
			expectCancelCallCount: 1,
			expectRemainingSubs:   3,
			expectErrorCount:      1,
		},
		{
			name:                  "TwoSubscriptionsRemovedWithError",
			intervals:             []time.Duration{2},
			cancelError:           true,
			expectCancelCallCount: 2,
			expectRemainingSubs:   3,
			expectErrorCount:      2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedSubscription := &SubscriptionMock{
				CancelFunc: func(ctx context.Context) error {
					if tc.cancelError {
						return testutils.ErrTesting
					}
					return nil
				},
			}
			m := NewMonitor(&Config{}, &ClientProviderMock{})
			m.AddSubscription(1, mockedSubscription)
			m.AddSubscription(2, &SubscriptionMock{})
			m.AddSubscription(3, mockedSubscription)

			errs := m.Purge(context.Background(), tc.intervals)

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

	m := NewMonitor(&Config{}, mockedClientProvider)
	var mockedSubscriptions [5]*SubscriptionMock
	for i := range mockedSubscriptions {
		mockedSubscription := &SubscriptionMock{
			CancelFunc: func(ctx context.Context) error {
				if len(mockedClientProvider.CloseWithContextCalls()) > 0 {
					t.Errorf("client has been closed before subscription cancel call")
				}
				return testutils.ErrTesting
			},
		}
		mockedSubscriptions[i] = mockedSubscription
		m.AddSubscription(time.Duration(i+1)*time.Second, mockedSubscription)
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
	m := NewMonitor(&Config{}, mockedClientProvider)

	if got, want := m.State(), opcua.ConnState(255); got != want {
		t.Errorf("State method: want %v, got %v", want, got)
	}
}
