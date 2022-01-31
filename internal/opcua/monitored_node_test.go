package opcua_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
)

func TestParseChannelSuccess(t *testing.T) {
	const channel = `opcua:s="node1"."node2";1800000`

	c, err := ParseChannel(channel)

	if msg := testutils.AssertError(t, err, false); msg != "" {
		t.Errorf("ParseChannel(): %s", msg)
	}
	if got, want := c.Node, `s="node1"."node2"`; got != want {
		t.Errorf("Node member: want %q, got %q", want, got)
	}
	if got, want := c.Interval, 30*time.Minute; got != want {
		t.Errorf("Interval member: want %v, got %v", want, got)
	}
	if got, want := c.Channel(), channel; got != want {
		t.Errorf("Channel() method: want %q, got %q", want, got)
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
			name:                  "TooManySemicolons",
			input:                 `opcua:ns=2;s="node1"."node2";1800000`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "IntervalParsingError",
			input:                 `opcua:s="node1"."node2";interval`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "NegativeInterval",
			input:                 `opcua:s="node1"."node2";-5000`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "IntervalTooBig",
			input:                 `opcua:s="node1"."node2";9223372036855`,
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
