package centrifugo_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
)

func TestParseChannel(t *testing.T) {
	cases := []struct {
		name                  string
		input                 string
		expectError           bool
		expectNotOpcUaChannel bool
		expectChannel         *Channel
	}{
		{
			name:                  "NoNamespace",
			input:                 "chan",
			expectError:           true,
			expectNotOpcUaChannel: true,
			expectChannel:         nil,
		},
		{
			name:                  "NotOpcUaNamespace",
			input:                 "ns:chan",
			expectError:           true,
			expectNotOpcUaChannel: true,
			expectChannel:         nil,
		},
		{
			name:                  "NodeUnescapeFailure",
			input:                 "opcua@1s:chan%2znode1%22.%22node2%22",
			expectError:           true,
			expectNotOpcUaChannel: false,
			expectChannel:         nil,
		},
		{
			name:                  "MissingInterval",
			input:                 "opcua:chan",
			expectError:           true,
			expectNotOpcUaChannel: false,
			expectChannel:         nil,
		},
		{
			name:                  "WrongInterval",
			input:                 "opcua@interval:chan",
			expectError:           true,
			expectNotOpcUaChannel: false,
			expectChannel:         nil,
		},
		{
			name:                  "Success",
			input:                 "opcua@30m:%22node1%22.%22node2%22",
			expectError:           false,
			expectNotOpcUaChannel: false,
			expectChannel:         &Channel{Node: `"node1"."node2"`, Interval: 30 * time.Minute},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := ParseChannel(tc.input)

			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if tc.expectError && err == nil {
				t.Errorf("expected an error, got nil")
			}
			if !tc.expectNotOpcUaChannel && errors.Is(err, ErrNotOpcUaChannel) {
				t.Errorf("unexpected ErrNotOpcUaChannel")
			}
			if tc.expectNotOpcUaChannel && !errors.Is(err, ErrNotOpcUaChannel) {
				t.Errorf("expected ErrNotOpcUaChannel, got %v", err)
			}
			if got, want := c, tc.expectChannel; !reflect.DeepEqual(got, want) {
				t.Errorf("want %#v, got %#v", want, got)
			}
		})
	}
}
