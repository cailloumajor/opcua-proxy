package opcua_test

import (
	"errors"
	"testing"

	. "github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
)

func TestParseChannelSuccess(t *testing.T) {
	const channel = `opcua:1800000`

	c, err := ParseChannel(channel)

	if msg := testutils.AssertError(t, err, false); msg != "" {
		t.Errorf("ParseChannel(): %s", msg)
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
			input:                 `30000`,
			expectNotOpcUaChannel: true,
		},
		{
			name:                  "NotOpcUaNamespace",
			input:                 `ns:30000`,
			expectNotOpcUaChannel: true,
		},
		{
			name:                  "IntervalParsingError",
			input:                 `opcua:interval`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "NegativeInterval",
			input:                 `opcua:-5000`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "IntervalTooBig",
			input:                 `opcua:9223372036855`,
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
