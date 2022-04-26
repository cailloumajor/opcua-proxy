package centrifugo_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
)

func TestParseChannelSuccess(t *testing.T) {
	const channel = `ns:my_tags@1800000`

	c, err := ParseChannel(channel, "ns")

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
		name                 string
		input                string
		expectIgnoredChannel bool
	}{
		{
			name:                 "MissingNamespace",
			input:                "my_tags@1800000",
			expectIgnoredChannel: true,
		},
		{
			name:                 "NotExpectedNamespace",
			input:                "otherns:my_tags@1800000",
			expectIgnoredChannel: true,
		},
		{
			name:                 "NoInterval",
			input:                "ns:my_tags",
			expectIgnoredChannel: false,
		},
		{
			name:                 "EmptyName",
			input:                "ns:@1800000",
			expectIgnoredChannel: false,
		},
		{
			name:                 "IntervalParsingError",
			input:                `ns:my_tags@interval`,
			expectIgnoredChannel: false,
		},
		{
			name:                 "NegativeInterval",
			input:                `ns:my_tags@-1800000`,
			expectIgnoredChannel: false,
		},
		{
			name:                 "IntervalTooBig",
			input:                `ns:my_tags@9223372036855`,
			expectIgnoredChannel: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseChannel(tc.input, "ns")

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
			}
			if got, want := errors.Is(err, ErrIgnoredChannel), tc.expectIgnoredChannel; got != want {
				t.Errorf("ignored namespace error returned: want %v, got %v", want, got)
			}
		})
	}
}
