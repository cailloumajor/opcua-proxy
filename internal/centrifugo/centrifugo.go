package centrifugo

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type sentinelError string

func (e sentinelError) Error() string {
	return string(e)
}

// ErrNotOpcUaChannel is issued when the channel is not suitable for OPC-UA.
const ErrNotOpcUaChannel = sentinelError("not an OPC-UA suitable channel")

// Channel represents a Centrifugo channel suitable for OPC-UA.
type Channel struct {
	Node     string        // OPC-UA node to monitor
	Interval time.Duration // Subscription interval
}

// ParseChannel parses a Centrifugo channel name into a channel structure.
func ParseChannel(s string) (*Channel, error) {
	split := strings.SplitN(s, ":", 2)
	if len(split) < 2 || !strings.HasPrefix(split[0], "opcua") {
		return nil, ErrNotOpcUaChannel
	}

	node, err := url.PathUnescape(split[1])
	if err != nil {
		return nil, fmt.Errorf("error unescaping node: %w", err)
	}

	split = strings.SplitN(split[0], "@", 2)
	if len(split) < 2 {
		return nil, fmt.Errorf("bad channel namespace format")
	}

	interval, err := time.ParseDuration(split[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing interval: %w", err)
	}

	return &Channel{Node: node, Interval: interval}, nil
}
