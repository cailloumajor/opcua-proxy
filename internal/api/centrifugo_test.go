package api_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/api"
	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
)

func TestCentrifugoSubscribeService(t *testing.T) {
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
			expectContentType: "application/problem+json",
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
			mockedChannelParser := &CentrifugoChannelParserMock{
				ParseChannelFunc: func(s, namespace string) (*centrifugo.Channel, error) {
					if tc.channelParseError != nil {
						return nil, tc.channelParseError
					}
					return &centrifugo.Channel{}, nil
				},
			}
			mockedOpcUaSubscriber := &OpcUaSubscriberMock{
				SubscribeFunc: func(ctx context.Context, nsURI string, ch opcua.ChannelProvider, nodes []opcua.NodeIDProvider) error {
					if tc.subscribeError {
						return testutils.ErrTesting
					}
					return nil
				},
			}
			s := NewCentrifugoSubscribeService(mockedChannelParser, mockedOpcUaSubscriber, "")

			r := strings.NewReader(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/", r)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

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
