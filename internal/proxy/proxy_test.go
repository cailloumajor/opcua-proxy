package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/cailloumajor/opcua-centrifugo/internal/proxy"
	"github.com/gopcua/opcua"
)

func TestHealth(t *testing.T) {
	cases := []struct {
		name             string
		gotState         opcua.ConnState
		expectStatusCode int
	}{
		{
			name:             "OpcClientNotConnected",
			gotState:         opcua.Disconnected,
			expectStatusCode: 500,
		},
		{
			name:             "OK",
			gotState:         opcua.Connected,
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedMonitorProvider := &MonitorProviderMock{
				StateFunc: func() opcua.ConnState {
					return tc.gotState
				},
			}
			proxy := NewProxy(mockedMonitorProvider)
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
