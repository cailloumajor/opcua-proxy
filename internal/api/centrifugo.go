package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
)

//go:generate moq -out centrifugo_mocks_test.go . CentrifugoChannelParser OpcUaSubscriber

const subscribeTimeout = 5 * time.Second

// CentrifugoChannelParser is a consumer contract modelling a Centrifugo channel name parser.
type CentrifugoChannelParser interface {
	ParseChannel(s, namespace string) (*centrifugo.Channel, error)
}

// OpcUaSubscriber is a consumer contract modelling an OPC-UA subscriber.
type OpcUaSubscriber interface {
	Subscribe(ctx context.Context, nsURI string, ch opcua.ChannelProvider, nodes []opcua.NodeIDProvider) error
}

// CentrifugoSubscribeService represents a service proxying Centrifugo subscriptions.
type CentrifugoSubscribeService struct {
	parser     CentrifugoChannelParser
	subscriber OpcUaSubscriber
	namespace  string
}

// NewCentrifugoSubscribeService creates a Centrifugo subscription proxy service.
func NewCentrifugoSubscribeService(p CentrifugoChannelParser, s OpcUaSubscriber, ns string) *CentrifugoSubscribeService {
	return &CentrifugoSubscribeService{p, s, ns}
}

func writeSuccessResponse(w io.Writer, msg string) {
	fmt.Fprintf(w, "{\"result\":{\"data\":{\"proxyMsg\":%q}}}\n", msg)
}

func writeErrorResponse(w io.Writer, code uint32, msg string) {
	fmt.Fprintf(w, "{\"error\":{\"code\":%d,\"message\":%q}}\n", code, msg)
}

// ServeHTTP implements http.Handler.
func (c *CentrifugoSubscribeService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sr struct {
		Channel string            `json:"channel"`
		Data    opcua.NodesObject `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
		ReplyProblemDetails(
			w,
			"request-content-decoding",
			"error decoding request content",
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set("content-type", "application/json")

	cch, err := c.parser.ParseChannel(sr.Channel, c.namespace)
	if errors.Is(err, centrifugo.ErrIgnoredChannel) {
		writeSuccessResponse(w, "ignored channel")
		return
	}
	if err != nil {
		msg := fmt.Sprintf("bad channel format: %v", err)
		writeErrorResponse(w, 1000, msg)
		return
	}

	nip := make([]opcua.NodeIDProvider, len(sr.Data.Nodes))
	for i := range sr.Data.Nodes {
		nip[i] = &sr.Data.Nodes[i]
	}

	ctx, cancel := context.WithTimeout(r.Context(), subscribeTimeout)
	defer cancel()

	if err := c.subscriber.Subscribe(ctx, sr.Data.NamespaceURI, cch, nip); err != nil {
		msg := fmt.Sprintf("error subscribing to OPC-UA data change: %v", err)
		writeErrorResponse(w, 1001, msg)
		return
	}

	writeSuccessResponse(w, "subscribed to OPC-UA data change")
}
