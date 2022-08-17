package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/lineprotocol"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/centrifugal/gocent/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gopcua "github.com/gopcua/opcua"
	"github.com/gorilla/mux"
	lp "github.com/influxdata/line-protocol/v2/lineprotocol"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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
	ParseChannel(s, namespace string) (*centrifugo.Channel, error)
}

// CentrifugoInfoProvider is a consumer contract modelling a Centrifugo server informations provider.
type CentrifugoInfoProvider interface {
	Info(ctx context.Context) (gocent.InfoResult, error)
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	sk := maps.Keys(m)
	slices.Sort(sk)
	return sk
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
	r.Methods("GET").Path("/influxdb-metrics").HandlerFunc(p.handleInfluxdbMetrics)
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

func (p *Proxy) handleInfluxdbMetrics(w http.ResponseWriter, r *http.Request) {
	handleErr := func(during string, err error, status int) {
		level.Info(p.logger).Log("during", during, "err", err)
		http.Error(w, http.StatusText(status), status)
	}

	if err := r.ParseForm(); err != nil {
		handleErr("parse request form", err, http.StatusInternalServerError)
		return
	}

	const measurementKey = "measurement"
	m := r.Form.Get(measurementKey)
	if m == "" {
		handleErr(
			"getting measurement query parameter",
			fmt.Errorf("missing measurement"),
			http.StatusBadRequest,
		)
		return
	}
	r.Form.Del(measurementKey)

	var enc lp.Encoder
	enc.StartLine(m)

	for _, fk := range sortedKeys(r.Form) {
		enc.AddTag(fk, r.Form.Get(fk))
	}

	vals, err := p.m.Read(r.Context())
	if err != nil {
		handleErr("nodes reading", err, http.StatusInternalServerError)
		return
	}

	for _, vk := range sortedKeys(vals.Values) {
		val, err := lineprotocol.NewValueFromVariant(vals.Values[vk])
		if err != nil {
			handleErr("converting variant to value", err, http.StatusInternalServerError)
			return
		}
		enc.AddField(vk, val)
	}

	enc.EndLine(vals.Timestamp.UTC())

	if err := enc.Err(); err != nil {
		handleErr("encoding line protocol", err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")

	if _, err := w.Write(enc.Bytes()); err != nil {
		handleErr("response data writing", err, http.StatusInternalServerError)
		return
	}
}

type subscribeRequest struct {
	Channel string            `json:"channel"`
	Data    opcua.NodesObject `json:"data"`
}

func writeSuccessResponse(w io.Writer, msg string) {
	fmt.Fprintf(w, "{\"result\":{\"data\":{\"proxyMsg\":%q}}}\n", msg)
}

func writeErrorResponse(w io.Writer, code uint32, msg string) {
	fmt.Fprintf(w, "{\"error\":{\"code\":%d,\"message\":%q}}\n", code, msg)
}

func (p *Proxy) handleCentrifugoSubscribe(w http.ResponseWriter, r *http.Request) {
	var sr subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
		msg := fmt.Sprintf("error decoding content: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	cch, err := p.cp.ParseChannel(sr.Channel, p.ns)
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

	if err := p.m.Subscribe(ctx, sr.Data.NamespaceURI, cch, nip); err != nil {
		msg := fmt.Sprintf("error subscribing to OPC-UA data change: %v", err)
		writeErrorResponse(w, 1001, msg)
		return
	}

	writeSuccessResponse(w, "subscribed to OPC-UA data change")
}
