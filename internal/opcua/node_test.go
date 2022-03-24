package opcua_test

import (
	"encoding/json"
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
	}{
		{
			name:         "UnknownType",
			json:         []byte("true"),
			expectError:  true,
			expectNodeID: "",
		},
		{
			name:         "NegativeNumber",
			json:         []byte("-42"),
			expectError:  true,
			expectNodeID: "",
		},
		{
			name:         "NotWholeNumber",
			json:         []byte("42.5"),
			expectError:  true,
			expectNodeID: "",
		},
		{
			name:         "OverflowingNumber",
			json:         []byte("4294967296"),
			expectError:  true,
			expectNodeID: "",
		},
		{
			name:         "GoodInteger",
			json:         []byte("42"),
			expectError:  false,
			expectNodeID: "ns=5;i=42",
		},
		{
			name:         "GoodString",
			json:         []byte("\"node1\""),
			expectError:  false,
			expectNodeID: "ns=5;s=node1",
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
		})
	}
}
