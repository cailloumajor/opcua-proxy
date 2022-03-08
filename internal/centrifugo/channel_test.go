package centrifugo_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
	"github.com/cailloumajor/opcua-centrifugo/internal/testutils"
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
		name                   string
		input                  string
		expectIgnoredNamespace bool
	}{
		{
			name:                   "MissingNamespace",
			input:                  "my_tags@1800000",
			expectIgnoredNamespace: true,
		},
		{
			name:                   "NotExpectedNamespace",
			input:                  "otherns:my_tags@1800000",
			expectIgnoredNamespace: true,
		},
		{
			name:                   "NoInterval",
			input:                  "ns:my_tags",
			expectIgnoredNamespace: false,
		},
		{
			name:                   "EmptyName",
			input:                  "ns:@1800000",
			expectIgnoredNamespace: false,
		},
		{
			name:                   "IntervalParsingError",
			input:                  `ns:my_tags@interval`,
			expectIgnoredNamespace: false,
		},
		{
			name:                   "NegativeInterval",
			input:                  `ns:my_tags@-1800000`,
			expectIgnoredNamespace: false,
		},
		{
			name:                   "IntervalTooBig",
			input:                  `ns:my_tags@9223372036855`,
			expectIgnoredNamespace: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseChannel(tc.input, "ns")

			if msg := testutils.AssertError(t, err, true); msg != "" {
				t.Errorf(msg)
			}
			if got, want := errors.Is(err, ErrIgnoredNamespace), tc.expectIgnoredNamespace; got != want {
				t.Errorf("ignored namespace error returned: want %v, got %v", want, got)
			}
		})
	}
}
