package opcua

func (m *Monitor) AddSubscription(interval PublishingInterval, sub Subscription) {
	m.subs[interval] = sub
}
