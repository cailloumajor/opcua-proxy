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
		ns                       string
		namespaceArrayError      bool
		namespaceNotFoundError   bool
		subscribeError           bool
		monitorError             bool
		monitoredItemCreateError bool
		expectSubCancelCalls     int
	}{
		{
			name:                     "NamespaceArrayError",
			ns:                       "ns0",
			namespaceArrayError:      true,
			subscribeError:           false,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "NamespaceNotFound",
			ns:                       "bad",
			namespaceArrayError:      false,
			subscribeError:           false,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "SubscribeError",
			ns:                       "ns0",
			namespaceArrayError:      false,
			subscribeError:           true,
			monitorError:             false,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "MonitorError",
			ns:                       "ns0",
			namespaceArrayError:      false,
			subscribeError:           false,
			monitorError:             true,
			monitoredItemCreateError: false,
			expectSubCancelCalls:     0,
		},
		{
			name:                     "MonitoredItemCreateError",
			ns:                       "ns0",
			namespaceArrayError:      false,
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
				NamespaceArrayWithContextFunc: func(ctx context.Context) ([]string, error) {
					if tc.namespaceArrayError {
						return nil, testutils.ErrTesting
					}
					return []string{"ns0"}, nil
				},
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error) {
					if tc.subscribeError {
						return nil, testutils.ErrTesting
					}
					return mockedSubscription, nil
				},
			}
			m := NewMonitor(&Config{}, mockedClientProvider)

			err := m.Subscribe(context.Background(), PublishingInterval(0), tc.ns, "node1", "node2", "node3")

			if got, want := len(mockedSubscription.CancelCalls()), tc.expectSubCancelCalls; got != want {
				t.Errorf("Cancel call count: want %d, got %d", want, got)
			}
			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}

func TestMonitorSubscribeSuccess(t *testing.T) {
	cases := []struct {
		name                   string
		interval               PublishingInterval
		nsURI                  string
		expectSubscribeCalled  bool
		expectedNamespaceIndex uint16
	}{
		{
			name:                   "SecondNamespace",
			interval:               0,
			nsURI:                  "ns1",
			expectSubscribeCalled:  false,
			expectedNamespaceIndex: 1,
		},
		{
			name:                   "ThirdNamespace",
			interval:               0,
			nsURI:                  "ns2",
			expectSubscribeCalled:  false,
			expectedNamespaceIndex: 2,
		},
		{
			name:                   "SubscribeCalled",
			interval:               1,
			nsURI:                  "ns0",
			expectSubscribeCalled:  true,
			expectedNamespaceIndex: 0,
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
				NamespaceArrayWithContextFunc: func(ctx context.Context) ([]string, error) {
					return []string{"ns0", "ns1", "ns2", "ns3"}, nil
				},
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error) {
					return mockedSubscription, nil
				},
			}
			m := NewMonitor(&Config{}, mockedClientProvider)
			m.AddSubscription(0, mockedSubscription)
			nodes := []string{"node1", "node2", "node3"}

			err := m.Subscribe(context.Background(), tc.interval, tc.nsURI, nodes...)

			if got, want := len(mockedClientProvider.NamespaceArrayWithContextCalls()), 1; got != want {
				t.Errorf("NamespaceArray call count: want %d, got %d", want, got)
			}
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
				if got, want := item.ItemToMonitor.NodeID.Namespace(), tc.expectedNamespaceIndex; got != want {
					t.Errorf("Monitor call, %q node namespace: want %d, got %d", nodes[i], want, got)
				}
				if got, want := item.ItemToMonitor.NodeID.StringID(), nodes[i]; got != want {
					t.Errorf("Monitor call, %q node string ID: want %q, got %q", nodes[i], want, got)
				}
				if got, want := item.RequestedParameters.ClientHandle, uint32(i); got != want {
					t.Errorf("Monitor call, %q node requested client handle: want %d, got %d", nodes[i], want, got)
				}
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
		m.AddSubscription(PublishingInterval(time.Duration(i+1)*time.Second), mockedSubscription)
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
