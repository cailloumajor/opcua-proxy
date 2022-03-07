package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gopcua/opcua"
)

//go:generate moq -out proxy_mocks_test.go . MonitorProvider ChannelProvider CentrifugoChannelParser

const subscribeTimeout = 5 * time.Second

// MonitorProvider is a consumer contract modelling an OPC-UA monitor.
type MonitorProvider interface {
	State() opcua.ConnState
	Subscribe(ctx context.Context, nsURI string, ch ChannelProvider, nodes []string) error
}

// ChannelProvider is a consumer contract modelling a Centrifugo channel.
type ChannelProvider interface {
	Name() string
	Interval() time.Duration
	fmt.Stringer
}

// CentrifugoChannelParser is a consumer contract modelling a Centrifugo channel parser.
type CentrifugoChannelParser interface {
	ParseChannel(s string) (ChannelProvider, error)
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

	http.Handler
}

// NewProxy creates and returns a ready to use proxy.
func NewProxy(m MonitorProvider, cp CentrifugoChannelParser) *Proxy {
	p := &Proxy{
		m:  m,
		cp: cp,
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

	switch {
	case p.m.State() != opcua.Connected:
		nok("OPC-UA client not connected")
	}
}

type subscribeRequest struct {
	Channel string `json:"channel"`
	Data    struct {
		NamespaceURI string   `json:"namespaceURI"`
		Nodes        []string `json:"nodes"`
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

func (p *Proxy) handleCentrifugoSubscribe(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	respondError := func(code uint32, msg string) {
		er := errorResponse{
			Error: errorContent{
				Code:    code,
				Message: msg,
			},
		}
		if err := enc.Encode(&er); err != nil {
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

	cch, err := p.cp.ParseChannel(sr.Channel)
	if err != nil {
		msg := fmt.Sprintf("bad channel format: %v", err)
		respondError(1000, msg)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), subscribeTimeout)
	defer cancel()

	if err := p.m.Subscribe(ctx, sr.Data.NamespaceURI, cch, sr.Data.Nodes); err != nil {
		msg := fmt.Sprintf("error subscribing to OPC-UA data change: %v", err)
		respondError(1001, msg)
		return
	}

	if err := enc.Encode(&successResponse{}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
