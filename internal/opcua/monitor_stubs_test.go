package opcua

import "time"

func (m *Monitor) AddSubscription(interval time.Duration, sub Subscription) {
	m.subs[interval] = sub
}
