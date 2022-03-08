package opcua

import (
	"context"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

// Subscription wraps an OPC-UA subscription.
type Subscription struct {
	inner *opcua.Subscription
}

// Cancel stub.
func (s *Subscription) Cancel(ctx context.Context) error {
	return s.inner.Cancel(ctx)
}

// ID returns the inner subscription ID.
func (s *Subscription) ID() uint32 {
	return s.inner.SubscriptionID
}

// MonitorWithContext stub.
func (s *Subscription) MonitorWithContext(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
	return s.inner.MonitorWithContext(ctx, ts, items...)
}
