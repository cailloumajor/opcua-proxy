package opcua_test

import (
	"encoding/json"
	"os"
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

func TestNodesObjectsFromFile(t *testing.T) {
	cases := []struct {
		name        string
		createFile  bool
		fileContent string
		expectError bool
	}{
		{
			name:        "NoFile",
			createFile:  false,
			fileContent: "",
			expectError: true,
		},
		{
			name:        "IncompleteFile",
			createFile:  true,
			fileContent: `[{"namespaceURI":"ns","nodes":[1,2]}`,
			expectError: true,
		},
		{
			name:        "Success",
			createFile:  true,
			fileContent: `[{"namespaceURI":"ns1","nodes":[1,2]},{"namespaceURI":"ns2","nodes":["3","4"]}]`,
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			fn := "nonexistent.json"
			if tc.createFile {
				f, err := os.CreateTemp(dir, "*.json")
				if err != nil {
					t.Fatalf("error creating the file: %v", err)
				}
				fn = f.Name()
				if _, err := f.Write([]byte(tc.fileContent)); err != nil {
					t.Fatalf("error writing the file: %v", err)
				}
				if err := f.Close(); err != nil {
					t.Fatalf("error closing the file: %v", err)
				}
			}

			no, err := NodesObjectsFromFile(fn)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if got := len(no); !tc.expectError && got <= 0 {
				t.Errorf("length of nodes objects slice: want it to be >0, got %d", got)
			}
		})
	}
}
