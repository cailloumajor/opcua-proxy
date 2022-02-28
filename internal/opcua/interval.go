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

// PublishingInterval represents a publishing interval.
type PublishingInterval time.Duration

// ParseChannel parses a Centrifugo channel into a publishing interval.
//
// See "Specifications" section in README.md for the format of the channel.
func ParseChannel(s string) (PublishingInterval, error) {
	if !strings.HasPrefix(s, channelPrefix) {
		return 0, ErrNotOpcUaChannel
	}

	cn := strings.TrimPrefix(s, channelPrefix)

	ms, err := strconv.ParseUint(cn, 10, 64)
	switch {
	case err != nil:
		return 0, fmt.Errorf("error parsing interval: %w", err)
	case ms > uint64(time.Duration(1<<63-1).Milliseconds()):
		return 0, fmt.Errorf("interval too big: %d", ms)
	}

	return PublishingInterval(time.Duration(ms) * time.Millisecond), nil
}

// Channel returns the Centrifugo channel name for this monitored node.
func (p PublishingInterval) Channel() string {
	return fmt.Sprint(channelPrefix, time.Duration(p).Milliseconds())
}
