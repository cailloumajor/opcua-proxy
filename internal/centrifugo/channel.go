package centrifugo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NsSeparator is the channel namespace boundary character.
const NsSeparator = ":"

const nameIntervalSeparator = "@"

type channelError string

func (c channelError) Error() string {
	return string(c)
}

// ErrIgnoredChannel is returned when the chennel is to be ignored.
const ErrIgnoredChannel = channelError("channel ignored")

// Channel represents a Centrifugo channel suitable for OPC-UA use.
type Channel struct {
	ns       string
	name     string
	interval time.Duration
}

// ParseChannel parses a Centrifugo channel and creates a Channel structure.
//
// See "Specifications" section in README.md for the format of the channel.
func ParseChannel(s, namespace string) (*Channel, error) {
	p := strings.SplitN(s, NsSeparator, 2)
	if len(p) != 2 {
		return nil, ErrIgnoredChannel
	}
	if p[0] != namespace || p[1] == HeartbeatChannel {
		return nil, ErrIgnoredChannel
	}
	ns, name := p[0], p[1]

	p = strings.SplitN(name, nameIntervalSeparator, 2)
	if len(p) != 2 {
		return nil, fmt.Errorf("bad channel name format: %q", name)
	}
	if p[0] == "" {
		return nil, fmt.Errorf("empty channel name: %q", p[0])
	}

	ms, err := strconv.ParseUint(p[1], 10, 64)
	switch {
	case err != nil:
		return nil, fmt.Errorf("error parsing interval: %w", err)
	case ms > uint64(time.Duration(1<<63-1).Milliseconds()):
		return nil, fmt.Errorf("interval too big: %d", ms)
	}

	return &Channel{
		ns:       ns,
		name:     p[0],
		interval: time.Duration(ms) * time.Millisecond,
	}, nil
}

// Name returns the channel name.
func (c *Channel) Name() string {
	return c.name
}

// Interval returns the channel interval.
func (c *Channel) Interval() time.Duration {
	return c.interval
}

// String returns the string representation of the channel.
func (c *Channel) String() string {
	return fmt.Sprint(c.ns, NsSeparator, c.name, nameIntervalSeparator, c.interval.Milliseconds())
}

// DefaultCentrifugoChannelParser is the default implementation of CentrifugoChannelParser.
type DefaultCentrifugoChannelParser struct{}

// ParseChannel implements CentrifugoChannelParser.
func (DefaultCentrifugoChannelParser) ParseChannel(s, namespace string) (*Channel, error) {
	return ParseChannel(s, namespace)
}
