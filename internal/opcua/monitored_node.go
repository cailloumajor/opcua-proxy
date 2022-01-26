package opcua

import (
	"fmt"
	"strings"
	"time"
)

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

// ParseChannel parses a Centrifugo channel name into an OPC-UA monitored node.
func ParseChannel(s string) (*MonitoredNode, error) {
	const ChannelPrefix = "opcua:"

	if !strings.HasPrefix(s, ChannelPrefix) {
		return nil, ErrNotOpcUaChannel
	}

	cn := strings.TrimPrefix(s, ChannelPrefix)

	p := strings.Split(cn, ";")
	switch {
	case len(p) < 2:
		return nil, fmt.Errorf("missing interval in channel name %q", cn)
	case len(p) > 2:
		return nil, fmt.Errorf("too many semicolons in channel name %q", cn)
	}

	i, err := time.ParseDuration(p[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing interval: %w", err)
	}

	return &MonitoredNode{
		Node:     p[0],
		Interval: i,
	}, nil
}
