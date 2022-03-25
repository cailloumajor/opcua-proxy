package proxy_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	. "github.com/cailloumajor/opcua-proxy/internal/proxy"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/centrifugal/gocent/v3"
	"github.com/go-kit/log"
	gopcua "github.com/gopcua/opcua"
)

func TestHealth(t *testing.T) {
	cases := []struct {
		name                string
		gotState            gopcua.ConnState
		centrifugoInfoError bool
		expectStatusCode    int
	}{
		{
			name:                "OpcClientNotConnected",
			gotState:            gopcua.Disconnected,
			centrifugoInfoError: false,
			expectStatusCode:    http.StatusInternalServerError,
		},
		{
			name:                "CentrifugoInfoError",
			gotState:            gopcua.Connected,
			centrifugoInfoError: true,
			expectStatusCode:    http.StatusInternalServerError,
		},
		{
			name:                "OK",
			gotState:            gopcua.Connected,
			centrifugoInfoError: false,
			expectStatusCode:    http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedMonitorProvider := &MonitorProviderMock{
				StateFunc: func() gopcua.ConnState {
					return tc.gotState
				},
			}
			mockedCentrifugoInfoProvider := &CentrifugoInfoProviderMock{
				InfoFunc: func(ctx context.Context) (gocent.InfoResult, error) {
					if tc.centrifugoInfoError {
						return gocent.InfoResult{}, testutils.ErrTesting
					}
					return gocent.InfoResult{
						Nodes: []gocent.NodeInfo{{}},
					}, nil
				},
			}
			proxy := NewProxy(log.NewNopLogger(), mockedMonitorProvider, &CentrifugoChannelParserMock{}, mockedCentrifugoInfoProvider, "")
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

func TestValues(t *testing.T) {
	cases := []struct {
		name              string
		querystring       string
		readError         bool
		expectStatusCode  int
		expectContentType string
		expectBody        string
		ignoreBody        bool
	}{
		{
			name:              "InvalidQueryString",
			querystring:       "tag1=val1&tag2=val2;",
			readError:         false,
			expectStatusCode:  http.StatusInternalServerError,
			expectContentType: "text/plain",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "ReadError",
			querystring:       "tag1=val1&tag2=val2",
			readError:         true,
			expectStatusCode:  http.StatusInternalServerError,
			expectContentType: "text/plain",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "Success",
			querystring:       "tag1=val1&tag2=val2",
			readError:         false,
			expectStatusCode:  http.StatusOK,
			expectContentType: "application/json",
			expectBody:        `{"timestamp":"2006-01-02T22:04:05Z","tags":{"tag1":"val1","tag2":"val2"},"fields":{"field1":37.2,"field2":"value"}}` + "\n",
			ignoreBody:        false,
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
						Values: map[string]interface{}{
							"field1": 37.2,
							"field2": "value",
						},
					}, nil
				},
			}
			p := NewProxy(log.NewNopLogger(), mockedMonitorProvider, &CentrifugoChannelParserMock{}, &CentrifugoInfoProviderMock{}, "")
			srv := httptest.NewServer(p)
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/values?" + tc.querystring)
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
			name:              "IgnoredNamespace",
			body:              "{}",
			channelParseError: centrifugo.ErrIgnoredNamespace,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectContentType: "application/json",
			expectBody:        "{\"result\":{}}\n",
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
			expectBody:        "{\"result\":{}}\n",
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
				ParseChannelFunc: func(s, namespace string) (opcua.ChannelProvider, error) {
					if tc.channelParseError != nil {
						return nil, tc.channelParseError
					}
					return &centrifugo.Channel{}, nil
				},
			}
			p := NewProxy(log.NewNopLogger(), mockedMonitorProvider, mockedChannelParser, &CentrifugoInfoProviderMock{}, "")
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
