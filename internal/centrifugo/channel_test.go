package centrifugo_test

import (
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
)

func TestParseChannelSuccess(t *testing.T) {
	const channel = `ns:my_tags@1800000`

	c, err := ParseChannel(channel)

	if msg := testutils.AssertError(t, err, false); msg != "" {
		t.Errorf(msg)
	}
	if got, want := c.Name(), "my_tags"; got != want {
		t.Errorf("Name method: want %q, got %q", want, got)
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
		name  string
		input string
	}{
		{
			name:  "MissingNamespace",
			input: "my_tags@1800000",
		},
		{
			name:  "NoInterval",
			input: "ns:my_tags",
		},
		{
			name:  "EmptyName",
			input: "ns:@1800000",
		},
		{
			name:  "IntervalParsingError",
			input: `ns:my_tags@interval`,
		},
		{
			name:  "NegativeInterval",
			input: `ns:my_tags@-1800000`,
		},
		{
			name:  "IntervalTooBig",
			input: `ns:my_tags@9223372036855`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseChannel(tc.input)

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}
