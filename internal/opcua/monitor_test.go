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
			m := NewMonitor(context.Background(), &Config{}, &ClientProviderMock{})
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
		CloseFunc: func() error {
			return testutils.ErrTesting
		},
	}

	m := NewMonitor(
		context.Background(),
		&Config{},
		mockedClientProvider,
	)
	var mockedSubscriptions [5]*SubscriptionMock
	for i := range mockedSubscriptions {
		mockedSubscription := &SubscriptionMock{
			CancelFunc: func(ctx context.Context) error {
				if len(mockedClientProvider.CloseCalls()) > 0 {
					t.Errorf("client has been closed before subscription cancel call")
				}
				return testutils.ErrTesting
			},
		}
		mockedSubscriptions[i] = mockedSubscription
		m.AddSubscription(PublishingInterval(time.Duration(i+1)*time.Second), mockedSubscription)
	}

	errs := m.Stop(context.Background())

	if got, want := len(mockedClientProvider.CloseCalls()), 1; got != want {
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
