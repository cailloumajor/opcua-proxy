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
	split := strings.SplitN(s, ":", 2)
	if len(split) < 2 || !strings.HasPrefix(split[0], "opcua") {
		return nil, ErrNotOpcUaChannel
	}

	split = strings.Split(split[1], ";")
	if len(split) < 2 || len(split) > 3 {
		return nil, fmt.Errorf("bad channel name format")
	}

	n := split[0]

	i, err := time.ParseDuration(split[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing interval: %w", err)
	}

	return &MonitoredNode{
		Node:     n,
		Interval: i,
	}, nil
}
