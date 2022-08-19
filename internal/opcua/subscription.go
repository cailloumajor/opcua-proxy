package opcua

import (
	"context"
	"fmt"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out subscription_mocks_test.go . Subscriber SubscriptionProvider

// Subscriber is a consumer contract modelling an OPC-UA subscriber.
type Subscriber interface {
	SubscribeWithContext(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (*opcua.Subscription, error)
}

// SubscriptionProvider is a consumer contract modelling an OPC-UA subscription.
type SubscriptionProvider interface {
	Cancel(ctx context.Context) error
	ID() uint32
	MonitorWithContext(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error)
}

// subscription wraps an OPC-UA subscription.
type subscription struct {
	*opcua.Subscription
}

// ID implements SubscriptionProvider.
func (s *subscription) ID() uint32 {
	return s.SubscriptionID
}

// Subscription wraps an subscription provider.
type Subscription struct {
	SubscriptionProvider
}

// NewSubscription creates a wrapped subscription provider.
func NewSubscription(sp SubscriptionProvider) *Subscription {
	return &Subscription{sp}
}

// SubscriptionManager is a manager for OPC-UA subscriptions.
type SubscriptionManager struct {
	subscriber Subscriber
}

// NewSubscriptionManager creates a subscriptions manager.
func NewSubscriptionManager(s Subscriber) *SubscriptionManager {
	return &SubscriptionManager{
		subscriber: s,
	}
}

// Create creates a subscription.
func (m *SubscriptionManager) Create(ctx context.Context, i time.Duration, nc chan<- *opcua.PublishNotificationData) (*Subscription, error) {
	p := &opcua.SubscriptionParameters{
		Interval: i,
	}

	s, err := m.subscriber.SubscribeWithContext(ctx, p, nc)
	if err != nil {
		return nil, fmt.Errorf("error creating subscription: %w", err)
	}

	return NewSubscription(&subscription{s}), nil
}
