package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/centrifugal/gocent/v3"
	gopcua "github.com/gopcua/opcua"
)

//go:generate moq -out proxy_mocks_test.go . MonitorProvider CentrifugoChannelParser CentrifugoInfoProvider

const subscribeTimeout = 5 * time.Second

// MonitorProvider is a consumer contract modelling an OPC-UA monitor.
type MonitorProvider interface {
	State() gopcua.ConnState
	Subscribe(ctx context.Context, nsURI string, ch opcua.ChannelProvider, nodes []opcua.NodeIDProvider) error
}

// CentrifugoChannelParser is a consumer contract modelling a Centrifugo channel parser.
type CentrifugoChannelParser interface {
	ParseChannel(s, namespace string) (opcua.ChannelProvider, error)
}

// CentrifugoInfoProvider is a consumer contract modelling a Centrifugo server informations provider.
type CentrifugoInfoProvider interface {
	Info(ctx context.Context) (gocent.InfoResult, error)
}

func methodHandler(m string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m {
			w.Header().Set("Allow", m)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Proxy handles requests for the service.
type Proxy struct {
	m  MonitorProvider
	cp CentrifugoChannelParser
	ci CentrifugoInfoProvider
	ns string // expected namespace for this instance

	http.Handler
}

// NewProxy creates and returns a ready to use proxy.
func NewProxy(m MonitorProvider, cp CentrifugoChannelParser, ci CentrifugoInfoProvider, ns string) *Proxy {
	p := &Proxy{
		m:  m,
		cp: cp,
		ci: ci,
		ns: ns,
	}

	mux := http.NewServeMux()
	mux.Handle("/health", methodHandler(http.MethodGet, http.HandlerFunc(p.handleHealth)))
	mux.Handle(
		"/centrifugo/subscribe",
		methodHandler(http.MethodPost, http.HandlerFunc(p.handleCentrifugoSubscribe)),
	)
	p.Handler = mux

	return p
}

func (p *Proxy) handleHealth(w http.ResponseWriter, r *http.Request) {
	nok := func(msg string) {
		http.Error(w, msg, http.StatusInternalServerError)
	}

	if p.m.State() != gopcua.Connected {
		nok("OPC-UA client not connected")
		return
	}

	if _, err := p.ci.Info(r.Context()); err != nil {
		nok(err.Error())
		return
	}
}

type subscribeRequest struct {
	Channel string `json:"channel"`
	Data    struct {
		NamespaceURI string       `json:"namespaceURI"`
		Nodes        []opcua.Node `json:"nodes"`
	} `json:"data"`
}

type successResponse struct {
	Result struct{} `json:"result"`
}

type errorContent struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorContent `json:"error"`
}

func newErrorResponse(code uint32, msg string) *errorResponse {
	return &errorResponse{
		Error: errorContent{
			Code:    code,
			Message: msg,
		},
	}
}

func (p *Proxy) handleCentrifugoSubscribe(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	respond := func(resp interface{}) {
		if err := enc.Encode(resp); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

	var sr subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
		msg := fmt.Sprintf("error decoding content: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	cch, err := p.cp.ParseChannel(sr.Channel, p.ns)
	if errors.Is(err, centrifugo.ErrIgnoredNamespace) {
		respond(&successResponse{})
		return
	}
	if err != nil {
		msg := fmt.Sprintf("bad channel format: %v", err)
		respond(newErrorResponse(1000, msg))
		return
	}

	nip := make([]opcua.NodeIDProvider, len(sr.Data.Nodes))
	for i := range sr.Data.Nodes {
		nip[i] = &sr.Data.Nodes[i]
	}

	ctx, cancel := context.WithTimeout(r.Context(), subscribeTimeout)
	defer cancel()

	if err := p.m.Subscribe(ctx, sr.Data.NamespaceURI, cch, nip); err != nil {
		msg := fmt.Sprintf("error subscribing to OPC-UA data change: %v", err)
		respond(newErrorResponse(1001, msg))
		return
	}

	respond(&successResponse{})
}

// DefaultCentrifugoChannelParser is the default implementation of CentrifugoChannelParser.
type DefaultCentrifugoChannelParser struct{}

// ParseChannel implements CentrifugoChannelParser.
func (DefaultCentrifugoChannelParser) ParseChannel(s, namespace string) (opcua.ChannelProvider, error) {
	return centrifugo.ParseChannel(s, namespace)
}
