package opcua

import (
	"github.com/gopcua/opcua"
)

func (m *Monitor) AddSubscription(channel string, sub SubscriptionProvider) {
	m.subs[channel] = sub
}

func (m *Monitor) PushNotification(n *opcua.PublishNotificationData) {
	m.notifyCh <- n
}

func (m *Monitor) Subs() map[string]SubscriptionProvider {
	return m.subs
}
