package centrifugo

import (
	"context"
	"fmt"

	"github.com/centrifugal/gocent/v3"
	"github.com/gopcua/opcua"
)

//go:generate moq -out heartbeat_mocks_test.go . StateProvider Publisher
//go:generate stringer -linecomment -trimprefix status -type status

// HeartbeatChannel is theCentrifugo channel name for heartbeat messages.
const HeartbeatChannel = "heartbeat"

// StateProvider is a consumer contract modelling an OPC-UA client.
type StateProvider interface {
	State() opcua.ConnState
}

// Publisher is a consumer contract modelling a Centrifugo publisher.
type Publisher interface {
	Publish(ctx context.Context, channel string, data []byte, opts ...gocent.PublishOption) (gocent.PublishResult, error)
}

type status uint8

const (
	statusOPCUaNotConnected status = iota // OPC-UA not connected
	statusOpcUaConnected                  // OPC-UA connected
)

// PublishStatus publishes the status of the service to Centrifugo heartbeat channel.
func PublishStatus(ctx context.Context, ns string, s StateProvider, p Publisher) error {
	var st status
	if s.State() == opcua.Connected {
		st = statusOpcUaConnected
	} else {
		st = statusOPCUaNotConnected
	}

	ch := ns + NsSeparator + HeartbeatChannel
	d := fmt.Sprintf("{\"status\":%d,\"description\":%q}", st, st)
	_, err := p.Publish(ctx, ch, []byte(d))

	return err
}
