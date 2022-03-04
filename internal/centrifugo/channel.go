package centrifugo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const nsSeparator = ":"
const channelPrefix = "opcua@"

type sentinelError string

func (e sentinelError) Error() string {
	return string(e)
}

// ErrNotOpcUaChannel is issued when the channel is not suitable for OPC-UA.
const ErrNotOpcUaChannel = sentinelError("not an OPC-UA suitable channel")

// Channel represents a Centrifugo channel suitable for OPC-UA use.
type Channel struct {
	ns       string
	interval time.Duration
}

// ParseChannel parses a Centrifugo channel and creates a Channel structure.
//
// It is expected that channel namespace is always present.
//
// See "Specifications" section in README.md for the format of the channel.
func ParseChannel(s string) (*Channel, error) {
	p := strings.SplitN(s, nsSeparator, 2)
	if len(p) < 2 {
		return nil, fmt.Errorf("missing namespace in %q channel", s)
	}

	if !strings.HasPrefix(p[1], channelPrefix) {
		return nil, ErrNotOpcUaChannel
	}

	i := strings.TrimPrefix(p[1], channelPrefix)

	ms, err := strconv.ParseUint(i, 10, 64)
	switch {
	case err != nil:
		return nil, fmt.Errorf("error parsing interval: %w", err)
	case ms > uint64(time.Duration(1<<63-1).Milliseconds()):
		return nil, fmt.Errorf("interval too big: %d", ms)
	}

	return &Channel{
		ns:       p[0],
		interval: time.Duration(ms) * time.Millisecond,
	}, nil
}

// Interval returns the channel interval.
func (c *Channel) Interval() time.Duration {
	return c.interval
}

// String returns the string representation of the channel.
func (c *Channel) String() string {
	return fmt.Sprint(c.ns, nsSeparator, channelPrefix, c.interval.Milliseconds())
}
