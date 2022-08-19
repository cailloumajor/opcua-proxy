package opcua

import (
	"github.com/gopcua/opcua"
)

func (m *Monitor) SetDummySub(s SubscriptionProvider) {
	m.dummySub = NewSubscription(s)
}

func (m *Monitor) InitDummyChannel() {
	m.dummyNotifCh = make(chan *opcua.PublishNotificationData)
}

func (m *Monitor) DummyChannel() chan *opcua.PublishNotificationData {
	return m.dummyNotifCh
}

func (m *Monitor) AddSubscription(channel string, sub SubscriptionProvider) {
	m.subs[channel] = &Subscription{sub}
}

func (m *Monitor) PushNotification(n *opcua.PublishNotificationData) {
	m.notifyCh <- n
}

func (m *Monitor) Subs() map[string]*Subscription {
	return m.subs
}
