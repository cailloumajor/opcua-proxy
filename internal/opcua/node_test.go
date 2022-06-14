package opcua_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
)

func TestNode(t *testing.T) {
	cases := []struct {
		name         string
		json         []byte
		expectError  bool
		expectNodeID string
		expectString string
	}{
		{
			name:         "UnknownType",
			json:         []byte("true"),
			expectError:  true,
			expectNodeID: "",
			expectString: "",
		},
		{
			name:         "NegativeNumber",
			json:         []byte("-42"),
			expectError:  true,
			expectNodeID: "",
			expectString: "",
		},
		{
			name:         "NotWholeNumber",
			json:         []byte("42.5"),
			expectError:  true,
			expectNodeID: "",
			expectString: "",
		},
		{
			name:         "OverflowingNumber",
			json:         []byte("4294967296"),
			expectError:  true,
			expectNodeID: "",
			expectString: "",
		},
		{
			name:         "GoodInteger",
			json:         []byte("42"),
			expectError:  false,
			expectNodeID: "ns=5;i=42",
			expectString: "42",
		},
		{
			name:         "GoodString",
			json:         []byte("\"node1\""),
			expectError:  false,
			expectNodeID: "ns=5;s=node1",
			expectString: "node1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var n Node
			err := json.Unmarshal(tc.json, &n)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := n.NodeID(5).String(), tc.expectNodeID; got != want {
				t.Errorf("NodeID method: want %#v, got %#v", want, got)
			}
			if got, want := n.String(), tc.expectString; got != want {
				t.Errorf("String method: want %q, got %q", want, got)
			}
		})
	}
}

func TestNodesObjectsFromURL(t *testing.T) {
	cases := []struct {
		name            string
		newRequestError bool
		requestError    bool
		httpError       bool
		content         string
		expectError     bool
	}{
		{
			name:            "NewRequestError",
			newRequestError: true,
			requestError:    false,
			httpError:       false,
			content:         "",
			expectError:     true,
		},
		{
			name:            "RequestError",
			newRequestError: false,
			requestError:    true,
			httpError:       false,
			content:         "",
			expectError:     true,
		},
		{
			name:            "HttpError",
			newRequestError: false,
			requestError:    false,
			httpError:       true,
			content:         "",
			expectError:     true,
		},
		{
			name:            "DecodeError",
			newRequestError: false,
			requestError:    false,
			httpError:       false,
			content:         `[{"namespaceURI":"ns","nodes":[1,2]}`,
			expectError:     true,
		},
		{
			name:            "Success",
			newRequestError: false,
			requestError:    false,
			httpError:       false,
			content:         `[{"namespaceURI":"ns1","nodes":[1,2]},{"namespaceURI":"ns2","nodes":["3","4"]}]`,
			expectError:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.httpError {
					http.Error(w, "testing error", http.StatusInternalServerError)
				} else {
					fmt.Fprintln(w, tc.content)
				}
			}))
			if tc.requestError {
				ts.Close()
			}
			defer ts.Close()

			var ctx context.Context
			if !tc.newRequestError {
				ctx = context.Background()
			}

			no, err := NodesObjectsFromURL(ctx, ts.URL)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if got := len(no); !tc.expectError && got <= 0 {
				t.Errorf("length of nodes objects slice: want it to be >0, got %d", got)
			}
		})
	}
}
