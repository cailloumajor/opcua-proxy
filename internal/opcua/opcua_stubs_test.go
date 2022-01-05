package opcua

func NewTestMonitor(c Client, m NodeMonitor) *Monitor {
	return &Monitor{client: c, nodeMonitor: m}
}
