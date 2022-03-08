package proxy_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
	"github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	. "github.com/cailloumajor/opcua-centrifugo/internal/proxy"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
	gopcua "github.com/gopcua/opcua"
)

func TestHealth(t *testing.T) {
	cases := []struct {
		name             string
		gotState         gopcua.ConnState
		expectStatusCode int
	}{
		{
			name:             "OpcClientNotConnected",
			gotState:         gopcua.Disconnected,
			expectStatusCode: 500,
		},
		{
			name:             "OK",
			gotState:         gopcua.Connected,
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedMonitorProvider := &MonitorProviderMock{
				StateFunc: func() gopcua.ConnState {
					return tc.gotState
				},
			}
			proxy := NewProxy(mockedMonitorProvider, &CentrifugoChannelParserMock{}, "")
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

func TestCentrifugoSubscribe(t *testing.T) {
	cases := []struct {
		name              string
		body              string
		channelParseError error
		subscribeError    bool
		expectStatusCode  int
		expectBody        string
		ignoreBody        bool
	}{
		{
			name:              "JsonDecodeError",
			body:              "[",
			channelParseError: nil,
			subscribeError:    false,
			expectStatusCode:  http.StatusBadRequest,
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "IgnoredNamespace",
			body:              "{}",
			channelParseError: centrifugo.ErrIgnoredNamespace,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectBody:        "{\"result\":{}}\n",
			ignoreBody:        false,
		},
		{
			name:              "ChannelParseError",
			body:              "{}",
			channelParseError: testutils.ErrTesting,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectBody:        "{\"error\":{\"code\":1000,\"message\":\"bad channel format: general error for testing\"}}\n",
			ignoreBody:        false,
		},
		{
			name:              "SubscribeError",
			body:              "{}",
			channelParseError: nil,
			subscribeError:    true,
			expectStatusCode:  http.StatusOK,
			expectBody:        "{\"error\":{\"code\":1001,\"message\":\"error subscribing to OPC-UA data change: general error for testing\"}}\n",
			ignoreBody:        false,
		},
		{
			name:              "Subscribed",
			body:              "{}",
			channelParseError: nil,
			subscribeError:    false,
			expectStatusCode:  http.StatusOK,
			expectBody:        "{\"result\":{}}\n",
			ignoreBody:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedMonitorProvider := &MonitorProviderMock{
				SubscribeFunc: func(ctx context.Context, nsURI string, ch opcua.ChannelProvider, nodes []string) error {
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
			p := NewProxy(mockedMonitorProvider, mockedChannelParser, "")
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
			if got, want := string(body), tc.expectBody; !tc.ignoreBody && got != want {
				t.Errorf("body: got %q, want %q", got, want)
			}
		})
	}
}
