package opcua_test

import (
	"context"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua"
)

func TestSubscriptionManagerCreate(t *testing.T) {
	cases := []struct {
		name           string
		subscribeError bool
		expectError    bool
	}{
		{
			name:           "SubscribeError",
			subscribeError: true,
			expectError:    true,
		},
		{
			name:           "Success",
			subscribeError: false,
			expectError:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedSubscriber := &SubscriberMock{
				SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error) {
					if tc.subscribeError {
						return nil, testutils.ErrTesting
					}
					return &opcua.Subscription{SubscriptionID: 4242}, nil
				},
			}
			m := NewSubscriptionManager(mockedSubscriber)
			ch := make(chan *opcua.PublishNotificationData, 1)

			s, err := m.Create(context.Background(), 42*time.Second, ch)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Error(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := mockedSubscriber.SubscribeWithContextCalls()[0].Params.Interval, 42*time.Second; got != want {
				t.Errorf("subscription interval: want %v, got %v", want, got)
			}
			p := &opcua.PublishNotificationData{}
			ch <- p
			select {
			case <-time.After(100 * time.Millisecond):
				t.Error("timeout getting the sentinel from the channel")
			case g := <-ch:
				if g != p {
					t.Error("unexpected sentinel")
				}
			}
			if got, want := s.ID(), uint32(4242); got != want {
				t.Errorf("subscription ID(): want %d, got %d", want, got)
			}
		})
	}
}
