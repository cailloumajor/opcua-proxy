package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/cailloumajor/opcua-proxy/internal/lineprotocol"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
)

//go:generate moq -out influxdb_metrics_mocks_test.go . OpcUaReader LineProtocolBuilder

// OpcUaReader is a consumer contract modelling an OPC-UA nodes values reader.
type OpcUaReader interface {
	Read(ctx context.Context) (*opcua.ReadValues, error)
}

// LineProtocolBuilder is a consumer contract modelling an InfluxDB line protocol builder.
type LineProtocolBuilder interface {
	Build(w io.Writer, measurement string, tags map[string]string, fields map[string]lineprotocol.VariantProvider, ts time.Time) error
}

// InfluxDbMetricsService represents a service providing InfluxDB metrics.
type InfluxDbMetricsService struct {
	reader  OpcUaReader
	builder LineProtocolBuilder
}

// NewInfluxDbMetricsService creates an InfluxDB metrics service.
func NewInfluxDbMetricsService(r OpcUaReader, b LineProtocolBuilder) *InfluxDbMetricsService {
	return &InfluxDbMetricsService{r, b}
}

// ServeHTTP implements http.Handler.
func (i *InfluxDbMetricsService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ReplyProblemDetails(
			w,
			"parse-request",
			"error parsing request form",
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	const measurementKey = "measurement"
	m := r.Form.Get(measurementKey)
	if m == "" {
		ReplyProblemDetails(
			w,
			"missing-measurement",
			"measurement query parameter not found",
			"",
			http.StatusBadRequest,
		)
		return
	}
	r.Form.Del(measurementKey)

	t := make(map[string]string)
	for k := range r.Form {
		t[k] = r.Form.Get(k)
	}

	vals, err := i.reader.Read(r.Context())
	if err != nil {
		ReplyProblemDetails(
			w,
			"nodes-reading",
			"error reading OPC-UA nodes",
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	f := make(map[string]lineprotocol.VariantProvider)
	for k, v := range vals.Values {
		f[k] = v
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")

	if err := i.builder.Build(w, m, t, f, vals.Timestamp); err != nil {
		ReplyProblemDetails(
			w,
			"line-protocol-encoding",
			"error encoding line protocol",
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
}
