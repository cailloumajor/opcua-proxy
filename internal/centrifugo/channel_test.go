package centrifugo_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
)

func TestParseChannelSuccess(t *testing.T) {
	const channel = `ns:opcua@1800000`

	c, err := ParseChannel(channel)

	if msg := testutils.AssertError(t, err, false); msg != "" {
		t.Errorf(msg)
	}
	if got, want := c.Interval(), 30*time.Minute; got != want {
		t.Errorf("Interval method: want %v, got %v", want, got)
	}
	if got, want := c.String(), channel; got != want {
		t.Errorf("String method: want %q, got %q", want, got)
	}
}

func TestParseChannelError(t *testing.T) {
	cases := []struct {
		name                  string
		input                 string
		expectNotOpcUaChannel bool
	}{
		{
			name:                  "MissingNamespace",
			input:                 "opcua@1800000",
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "NotOpcUaChannel",
			input:                 `ns:1800000`,
			expectNotOpcUaChannel: true,
		},
		{
			name:                  "IntervalParsingError",
			input:                 `ns:opcua@interval`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "NegativeInterval",
			input:                 `ns:opcua@-1800000`,
			expectNotOpcUaChannel: false,
		},
		{
			name:                  "IntervalTooBig",
			input:                 `ns:opcua@9223372036855`,
			expectNotOpcUaChannel: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseChannel(tc.input)

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
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
