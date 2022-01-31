package opcua

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const channelPrefix = "opcua:"

type sentinelError string

func (e sentinelError) Error() string {
	return string(e)
}

// ErrNotOpcUaChannel is issued when the channel is not suitable for OPC-UA.
const ErrNotOpcUaChannel = sentinelError("not an OPC-UA suitable channel")

// MonitoredNode represents a monitored OPC-UA node.
type MonitoredNode struct {
	Node     string        // Node identifier without namespace
	Interval time.Duration // Subscription interval
}

// ParseChannel parses a Centrifugo channel into an OPC-UA monitored node.
//
// See "Specifications" section in README.md for the format of the channel.
func ParseChannel(s string) (*MonitoredNode, error) {
	if !strings.HasPrefix(s, channelPrefix) {
		return nil, ErrNotOpcUaChannel
	}

	cn := strings.TrimPrefix(s, channelPrefix)

	p := strings.Split(cn, ";")
	switch {
	case len(p) < 2:
		return nil, fmt.Errorf("missing interval in channel name %q", cn)
	case len(p) > 2:
		return nil, fmt.Errorf("too many semicolons in channel name %q", cn)
	}

	ms, err := strconv.ParseUint(p[1], 10, 64)
	switch {
	case err != nil:
		return nil, fmt.Errorf("error parsing interval: %w", err)
	case ms > uint64(time.Duration(1<<63-1).Milliseconds()):
		return nil, fmt.Errorf("interval too big: %d", ms)
	}

	return &MonitoredNode{
		Node:     p[0],
		Interval: time.Duration(ms) * time.Millisecond,
	}, nil
}

// Channel returns the Centrifugo channel name for this monitored node.
func (m *MonitoredNode) Channel() string {
	return fmt.Sprint(channelPrefix, m.Node, ";", m.Interval.Milliseconds())
}
