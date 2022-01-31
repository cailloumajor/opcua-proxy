package opcua_test

import (
	"context"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
)

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
		m.AddSubscription(time.Duration(i+1)*time.Second, mockedSubscription)
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
