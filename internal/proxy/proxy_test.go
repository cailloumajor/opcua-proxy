package proxy_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/lineprotocol"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	. "github.com/cailloumajor/opcua-proxy/internal/proxy"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/go-kit/log"
	"github.com/gopcua/opcua/ua"
)

func TestHealth(t *testing.T) {
	cases := []struct {
		name              string
		monitorHealthy    bool
		centrifugoHealthy bool
		expectStatusCode  int
	}{
		{
			name:              "MonitorUnhealthy",
			monitorHealthy:    false,
			centrifugoHealthy: true,
			expectStatusCode:  http.StatusInternalServerError,
		},
		{
			name:              "CentrifugoUnhealthy",
			monitorHealthy:    true,
			centrifugoHealthy: false,
			expectStatusCode:  http.StatusInternalServerError,
		},
		{
			name:              "OK",
			monitorHealthy:    true,
			centrifugoHealthy: true,
			expectStatusCode:  http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedCentrifugoHealther := &HealtherMock{
				HealthFunc: func(ctx context.Context) (bool, string) {
					if !tc.centrifugoHealthy {
						return false, "ill"
					}
					return true, "well"
				},
			}
			mockedMonitorProvider := &MonitorProviderMock{
				HealthFunc: func(ctx context.Context) (bool, string) {
					if !tc.monitorHealthy {
						return false, "sick"
					}
					return true, "good"
				},
			}
			proxy := NewProxy(
				log.NewNopLogger(),
				mockedCentrifugoHealther,
				mockedMonitorProvider,
				&LineProtocolBuilderMock{},
				&CentrifugoChannelParserMock{},
				"",
			)
			srv := httptest.NewServer(proxy)
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/health")
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			if got, want := resp.StatusCode, tc.expectStatusCode; got != want {
				t.Errorf("status code: want %d, got %d", want, got)
			}
		})
	}
}

func TestInfluxdbMetrics(t *testing.T) {
	cases := []struct {
		name             string
		querystring      string
		readError        bool
		buildError       bool
		expectStatusCode int
		expectBody       string
		ignoreBody       bool
	}{
		{
			name:             "InvalidQueryString",
			querystring:      "measurement=meas&tag1=val1&othertag=otherval;",
			readError:        false,
			buildError:       false,
			expectStatusCode: http.StatusInternalServerError,
			expectBody:       "",
			ignoreBody:       true,
		},
		{
			name:             "MissingMeasurement",
			querystring:      "tag1=val1&othertag=otherval",
			readError:        false,
			buildError:       false,
			expectStatusCode: http.StatusBadRequest,
			expectBody:       "",
			ignoreBody:       true,
		},
		{
			name:             "ReadError",
			querystring:      "measurement=meas&tag1=val1&othertag=otherval",
			readError:        true,
			buildError:       false,
			expectStatusCode: http.StatusInternalServerError,
			expectBody:       "",
			ignoreBody:       true,
		},
		{
			name:             "BuildError",
			querystring:      "measurement=meas&tag1=val1&othertag=otherval",
			readError:        false,
			buildError:       true,
			expectStatusCode: http.StatusInternalServerError,
			expectBody:       "",
			ignoreBody:       true,
		},
		{
			name:             "Success",
			querystring:      "measurement=meas&tag1=val1&othertag=otherval",
			readError:        false,
			buildError:       false,
			expectStatusCode: http.StatusOK,
			expectBody:       "meas map[othertag:otherval tag1:val1] map[field1: field2:val] 2006-01-02 15:04:05 -0700 -0700",
			ignoreBody:       false,
		},
	}

	ts, err := time.Parse(time.Layout, time.Layout)
	if err != nil {
		t.Fatalf("error parsing time: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedMonitorProvider := &MonitorProviderMock{
				ReadFunc: func(ctx context.Context) (*opcua.ReadValues, error) {
					if tc.readError {
						return nil, testutils.ErrTesting
					}
					return &opcua.ReadValues{
						Timestamp: ts,
						Values: map[string]*ua.Variant{
							"field1": ua.MustVariant(37.2),
							"field2": ua.MustVariant("val"),
						},
					}, nil
				},
			}
			mockedLineProtocolBuilder := &LineProtocolBuilderMock{
				BuildFunc: func(w io.Writer, measurement string, tags map[string]string, fields map[string]lineprotocol.VariantProvider, ts time.Time) error {
					if tc.buildError {
						return testutils.ErrTesting
					}
					if _, err := fmt.Fprintf(w, "%v %v %v %v", measurement, tags, fields, ts); err != nil {
						t.Fatalf("error writing to ResponseWriter: %v", err)
					}
					return nil
				},
			}
			p := NewProxy(
				log.NewNopLogger(),
				&HealtherMock{},
				mockedMonitorProvider,
				mockedLineProtocolBuilder,
				&CentrifugoChannelParserMock{},
				"",
			)
			srv := httptest.NewServer(p)
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/influxdb-metrics?" + tc.querystring)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("error reading response body: %v", err)
			}
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("error closing response body: %v", err)
			}

			if got, want := resp.StatusCode, tc.expectStatusCode; got != want {
				t.Errorf("status code: want %d, got %d", want, got)
			}
			if got, want := resp.Header.Get("content-type"), "text/plain"; !strings.HasPrefix(got, want) {
				t.Errorf("content type: want %q, got %q", want, got)
			}
			if got, want := string(body), tc.expectBody; !tc.ignoreBody && got != want {
				t.Errorf("body: got %q, want %q", got, want)
			}
		})
	}
}

func TestCentrifugoSubscribe(t *testing.T) {
	cases := []struct {
		name              string
		body              string
		channelParseError error
		subscribeError    bool
		expectStatusCode  int
		expectContentType string
		expectBody        string
		ignoreBody        bool
	}{
		{
			name:              "JsonDecodeError",
			body:              "[",
			channelParseError: nil,
			subscribeError:    false,
			expectStatusCode:  http.StatusBadRequest,
			expectContentType: "text/plain",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "IgnoredChannel",
			body:              "{}",
			channelParseError: centrifugo.ErrIgnoredChannel,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectContentType: "application/json",
			expectBody:        "{\"result\":{\"data\":{\"proxyMsg\":\"ignored channel\"}}}\n",
			ignoreBody:        false,
		},
		{
			name:              "ChannelParseError",
			body:              "{}",
			channelParseError: testutils.ErrTesting,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectContentType: "application/json",
			expectBody:        "{\"error\":{\"code\":1000,\"message\":\"bad channel format: general error for testing\"}}\n",
			ignoreBody:        false,
		},
		{
			name:              "SubscribeError",
			body:              "{}",
			channelParseError: nil,
			subscribeError:    true,
			expectStatusCode:  http.StatusOK,
			expectContentType: "application/json",
			expectBody:        "{\"error\":{\"code\":1001,\"message\":\"error subscribing to OPC-UA data change: general error for testing\"}}\n",
			ignoreBody:        false,
		},
		{
			name:              "Subscribed",
			body:              `{"channel":"ch1","data":{"namespaceURI":"uri","nodes":[""]}}`,
			channelParseError: nil,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectContentType: "application/json",
			expectBody:        "{\"result\":{\"data\":{\"proxyMsg\":\"subscribed to OPC-UA data change\"}}}\n",
			ignoreBody:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedMonitorProvider := &MonitorProviderMock{
				SubscribeFunc: func(ctx context.Context, nsURI string, ch opcua.ChannelProvider, nodes []opcua.NodeIDProvider) error {
					if tc.subscribeError {
						return testutils.ErrTesting
					}
					return nil
				},
			}
			mockedChannelParser := &CentrifugoChannelParserMock{
				ParseChannelFunc: func(s, namespace string) (*centrifugo.Channel, error) {
					if tc.channelParseError != nil {
						return nil, tc.channelParseError
					}
					return &centrifugo.Channel{}, nil
				},
			}
			p := NewProxy(
				log.NewNopLogger(),
				&HealtherMock{},
				mockedMonitorProvider,
				&LineProtocolBuilderMock{},
				mockedChannelParser,
				"",
			)
			srv := httptest.NewServer(p)
			defer srv.Close()

			r := strings.NewReader(tc.body)
			resp, err := http.Post(srv.URL+"/centrifugo/subscribe", "application/json", r)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("error reading response body: %v", err)
			}
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("error closing response body: %v", err)
			}

			if got, want := resp.StatusCode, tc.expectStatusCode; got != want {
				t.Errorf("status code: want %d, got %d", want, got)
			}
			if got, want := resp.Header.Get("content-type"), tc.expectContentType; !strings.HasPrefix(got, want) {
				t.Errorf("content type: want %q, got %q", want, got)
			}
			if got, want := string(body), tc.expectBody; !tc.ignoreBody && got != want {
				t.Errorf("body: got %q, want %q", got, want)
			}
		})
	}
}
