package opcua

import (
	"time"

	"github.com/gopcua/opcua"
)

func (m *Monitor) AddSubscription(name string, interval time.Duration, sub SubscriptionProvider) {
	m.subs[subShape{name: name, interval: interval}] = sub
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

func (m *Monitor) Subs() map[subShape]SubscriptionProvider {
	return m.subs
}

func (m *Monitor) Chans() map[uint32]ChannelProvider {
	return m.chans
}

func (m *Monitor) AddChan(id uint32, c ChannelProvider) {
	m.chans[id] = c
}

func (m *Monitor) Items() map[uint32]string {
	return m.items
}
