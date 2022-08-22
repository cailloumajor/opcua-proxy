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
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
)

//go:generate moq -out proxy_mocks_test.go . Healther MonitorProvider LineProtocolBuilder CentrifugoChannelParser

const subscribeTimeout = 5 * time.Second

// Healther is a consumer contract modelling a health status provider.
type Healther interface {
	Health(ctx context.Context) (bool, string)
}

// MonitorProvider is a consumer contract modelling an OPC-UA monitor.
type MonitorProvider interface {
	Healther
	Read(ctx context.Context) (*opcua.ReadValues, error)
	Subscribe(ctx context.Context, nsURI string, ch opcua.ChannelProvider, nodes []opcua.NodeIDProvider) error
}

// LineProtocolBuilder is a consumer contract modelling an InfluxDB line protocol builder.
type LineProtocolBuilder interface {
	Build(w io.Writer, measurement string, tags map[string]string, fields map[string]lineprotocol.VariantProvider, ts time.Time) error
}

// CentrifugoChannelParser is a consumer contract modelling a Centrifugo channel parser.
type CentrifugoChannelParser interface {
	ParseChannel(s, namespace string) (*centrifugo.Channel, error)
}

func commonHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// Proxy handles requests for the service.
type Proxy struct {
	logger           log.Logger
	centrifugoHealth Healther
	monitor          MonitorProvider
	lpBuilder        LineProtocolBuilder
	parser           CentrifugoChannelParser
	namespace        string // expected namespace for this instance

	http.Handler
}

// NewProxy creates and returns a ready to use proxy.
func NewProxy(l log.Logger, ch Healther, m MonitorProvider, lp LineProtocolBuilder, cp CentrifugoChannelParser, ns string) *Proxy {
	p := &Proxy{
		logger:           l,
		centrifugoHealth: ch,
		monitor:          m,
		lpBuilder:        lp,
		parser:           cp,
		namespace:        ns,
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
	fail := func(module, reason string) {
		msg := fmt.Sprintf("unhealthy %v: %v", module, reason)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	if ok, msg := p.monitor.Health(context.Background()); !ok {
		fail("OPC-UA monitor", msg)
		return
	}

	if ok, msg := p.centrifugoHealth.Health(r.Context()); !ok {
		fail("Centrifugo", msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	t := make(map[string]string)
	for k := range r.Form {
		t[k] = r.Form.Get(k)
	}

	vals, err := p.monitor.Read(r.Context())
	if err != nil {
		handleErr("nodes reading", err, http.StatusInternalServerError)
		return
	}

	f := make(map[string]lineprotocol.VariantProvider)
	for k, v := range vals.Values {
		f[k] = v
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")

	if err := p.lpBuilder.Build(w, m, t, f, vals.Timestamp); err != nil {
		handleErr("line protocol builing", err, http.StatusInternalServerError)
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

	cch, err := p.parser.ParseChannel(sr.Channel, p.namespace)
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

	if err := p.monitor.Subscribe(ctx, sr.Data.NamespaceURI, cch, nip); err != nil {
		msg := fmt.Sprintf("error subscribing to OPC-UA data change: %v", err)
		writeErrorResponse(w, 1001, msg)
		return
	}

	writeSuccessResponse(w, "subscribed to OPC-UA data change")
}
