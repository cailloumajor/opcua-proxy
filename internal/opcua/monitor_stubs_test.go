package opcua

import (
	"github.com/gopcua/opcua"
)

func (m *Monitor) AddSubscription(interval PublishingInterval, sub Subscription) {
	m.subs[interval] = sub
}

func (m *Monitor) AddMonitoredItems(nodes ...string) {
	for _, n := range nodes {
		l := uint32(len(m.items))
		m.items[l] = n
	}
}

func (m *Monitor) PushNotification(n *opcua.PublishNotificationData) {
	m.notifyCh <- n
}
