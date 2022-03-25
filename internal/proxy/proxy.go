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
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gopcua "github.com/gopcua/opcua"
	"github.com/gorilla/mux"
)

//go:generate moq -out proxy_mocks_test.go . MonitorProvider CentrifugoChannelParser CentrifugoInfoProvider

const subscribeTimeout = 5 * time.Second

// MonitorProvider is a consumer contract modelling an OPC-UA monitor.
type MonitorProvider interface {
	Read(ctx context.Context) (*opcua.ReadValues, error)
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

func commonHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// Proxy handles requests for the service.
type Proxy struct {
	logger log.Logger
	m      MonitorProvider
	cp     CentrifugoChannelParser
	ci     CentrifugoInfoProvider
	ns     string // expected namespace for this instance

	http.Handler
}

// NewProxy creates and returns a ready to use proxy.
func NewProxy(l log.Logger, m MonitorProvider, cp CentrifugoChannelParser, ci CentrifugoInfoProvider, ns string) *Proxy {
	p := &Proxy{
		logger: l,
		m:      m,
		cp:     cp,
		ci:     ci,
		ns:     ns,
	}

	r := mux.NewRouter()
	r.Methods("GET").Path("/health").HandlerFunc(p.handleHealth)
	r.Methods("GET").Path("/values").HandlerFunc(p.handleValues)
	r.Methods("POST").Path("/centrifugo/subscribe").HandlerFunc(p.handleCentrifugoSubscribe)
	r.Use(commonHeadersMiddleware)
	p.Handler = r

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

type valuesResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

func (p *Proxy) handleValues(w http.ResponseWriter, r *http.Request) {
	handleErr := func(during string, err error) {
		level.Info(p.logger).Log("during", during, "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if err := r.ParseForm(); err != nil {
		handleErr("parse request form", err)
		return
	}

	resp := &valuesResponse{
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}

	for k := range r.Form {
		resp.Tags[k] = r.Form.Get(k)
	}

	vals, err := p.m.Read(r.Context())
	if err != nil {
		handleErr("nodes reading", err)
		return
	}

	resp.Timestamp = vals.Timestamp.UTC()
	resp.Fields = vals.Values

	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleErr("response encoding", err)
		return
	}
}

type subscribeRequest struct {
	Channel string            `json:"channel"`
	Data    opcua.NodesObject `json:"data"`
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
