package centrifugo_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
)

func TestParseChannelSuccess(t *testing.T) {
	c, err := ParseChannel(`opcua:s="node1"."node2";30m`)

	if msg := testutils.AssertError(t, err, false); msg != "" {
		t.Errorf("Node() method: %s", msg)
	}
	if got, want := c.Node(), `s="node1"."node2"`; got != want {
		t.Errorf("Node() method: want %q, got %q", want, got)
	}
	if got, want := c.Interval(), 30*time.Minute; got != want {
		t.Errorf("Interval() method: want %v, got %v", want, got)
	}
}

func TestParseChannelError(t *testing.T) {
	cases := []struct {
		name                  string
		input                 string
		expectNotOpcUaChannel bool
	}{
		{
			name:                  "NoNamespace",
			input:                 `s="node1"."node2"`,
			expectNotOpcUaChannel: true,
		},
		{
			name:                  "NotOpcUaNamespace",
			input:                 `ns:s="node1"."node2"`,
			expectNotOpcUaChannel: true,
		},
		{
			name:                  "MissingInterval",
			input:                 `opcua:s="node1"."node2"`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "WrongInterval",
			input:                 `opcua:s="node1"."node2";interval`,
			expectNotOpcUaChannel: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseChannel(tc.input)

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf("Node() method: %s", msg)
			}
			if !tc.expectNotOpcUaChannel && errors.Is(err, ErrNotOpcUaChannel) {
				t.Errorf("unexpected ErrNotOpcUaChannel")
			}
			if tc.expectNotOpcUaChannel && !errors.Is(err, ErrNotOpcUaChannel) {
				t.Errorf("expected ErrNotOpcUaChannel, got %v", err)
			}
		})
	}
}
