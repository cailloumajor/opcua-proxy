package centrifugo

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

// Channel represents a Centrifugo channel suitable for OPC-UA.
type Channel struct {
	node     string
	interval time.Duration
}

// ParseChannel parses a Centrifugo channel name into a channel structure.
func ParseChannel(s string) (*Channel, error) {
	split := strings.SplitN(s, ":", 2)
	if len(split) < 2 || !strings.HasPrefix(split[0], "opcua") {
		return nil, ErrNotOpcUaChannel
	}

	split = strings.Split(split[1], ";")
	if len(split) < 2 || len(split) > 3 {
		return nil, fmt.Errorf("bad channel name format")
	}

	node := split[0]

	interval, err := time.ParseDuration(split[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing interval: %w", err)
	}

	return &Channel{node: node, interval: interval}, nil
}

// Node returns the OPC-UA node of the channel.
func (c *Channel) Node() string {
	return c.node
}

// Interval returns the subscription interval of the channel.
func (c *Channel) Interval() time.Duration {
	return c.interval
}
