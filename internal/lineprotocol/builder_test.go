package lineprotocol_test

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-proxy/internal/lineprotocol"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, testutils.ErrTesting
}

func TestBuilderBuildImplementation(t *testing.T) {
	b := NewBuilder()

	sb := &strings.Builder{}

	_ = b.Build(sb, "m", nil, map[string]VariantProvider{"k": ua.MustVariant(byte(0))}, time.Time{})

	if got, want := sb.String(), "m k=0u\n"; got != want {
		t.Errorf("written data: want %q, got %q", want, got)
	}
}

func TestBuilderBuild(t *testing.T) {
	cases := []struct {
		name               string
		unsupportedVariant bool
		encoderError       bool
		writeError         bool
		expectError        bool
	}{
		{
			name:               "UnsupportedVariant",
			unsupportedVariant: true,
			encoderError:       false,
			writeError:         false,
			expectError:        true,
		},
		{
			name:               "EncodeError",
			unsupportedVariant: false,
			encoderError:       true,
			writeError:         false,
			expectError:        true,
		},
		{
			name:               "WriteError",
			unsupportedVariant: false,
			encoderError:       false,
			writeError:         true,
			expectError:        true,
		},
		{
			name:               "Success",
			unsupportedVariant: false,
			encoderError:       false,
			writeError:         false,
			expectError:        false,
		},
	}

	mockedUnsupportedVariant := &VariantProviderMock{
		TypeFunc: func() ua.TypeID { return ua.TypeIDNull },
	}
	mockedFloatVariant := &VariantProviderMock{
		TypeFunc:  func() ua.TypeID { return ua.TypeIDFloat },
		FloatFunc: func() float64 { return 37.2 },
	}
	mockedStringVariant := &VariantProviderMock{
		TypeFunc:   func() ua.TypeID { return ua.TypeIDString },
		StringFunc: func() string { return "value" },
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedEncoder := &EncoderMock{
				AddFieldFunc: func(key string, value lineprotocol.Value) {},
				AddTagFunc:   func(key, value string) {},
				BytesFunc:    func() []byte { return []byte("encoded content") },
				EndLineFunc:  func(t time.Time) {},
				ErrFunc: func() error {
					if tc.encoderError {
						return testutils.ErrTesting
					}
					return nil
				},
				ResetFunc:     func() {},
				StartLineFunc: func(measurement string) {},
			}
			mockedPooler := &PoolerMock{
				GetFunc: func() any { return mockedEncoder },
				PutFunc: func(x any) {},
			}

			b := NewMockedBuilder(mockedPooler)

			tags := map[string]string{"tag1": "val1", "othertag": "otherval"}
			fields := map[string]VariantProvider{"field1": mockedFloatVariant}
			if tc.unsupportedVariant {
				fields["anotherfield"] = mockedUnsupportedVariant
			} else {
				fields["anotherfield"] = mockedStringVariant
			}
			sb := &strings.Builder{}
			var w io.Writer
			if tc.writeError {
				w = &errorWriter{}
			} else {
				w = sb
			}
			ts := time.Now()
			err := b.Build(w, "meas", tags, fields, ts)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Error(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := len(mockedEncoder.ResetCalls()), 1; got != want {
				t.Errorf("Reset() call count: want %d, got %d", want, got)
			}
			if got, want := len(mockedEncoder.StartLineCalls()), 1; got != want {
				t.Errorf("StartLine() call count: want %d, got %d", want, got)
			}
			if got, want := mockedEncoder.StartLineCalls()[0].Measurement, "meas"; got != want {
				t.Errorf("StartLine() measurement argument: want %q, got %q", want, got)
			}
			expectedAddTagCalls := []struct {
				Key   string
				Value string
			}{
				{Key: "othertag", Value: "otherval"},
				{Key: "tag1", Value: "val1"},
			}
			if got, want := mockedEncoder.AddTagCalls(), expectedAddTagCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("AddTag() calls: want %v, got %v", want, got)
			}
			expectedAddFieldCalls := []struct {
				Key   string
				Value lineprotocol.Value
			}{
				{Key: "anotherfield", Value: lineprotocol.MustNewValue("value")},
				{Key: "field1", Value: lineprotocol.MustNewValue(37.2)},
			}
			if got, want := mockedEncoder.AddFieldCalls(), expectedAddFieldCalls; !reflect.DeepEqual(got, want) {
				t.Errorf("AddField() calls: want %v, got %v", want, got)
			}
			if got, want := len(mockedEncoder.EndLineCalls()), 1; got != want {
				t.Errorf("EndLine() call count: want %d, got %d", want, got)
			}
			if got, want := mockedEncoder.EndLineCalls()[0].T, ts; !got.Equal(want) {
				t.Errorf("EndLine() t argument: want %v, got %v", want, got)
			}
			if got, want := mockedEncoder.EndLineCalls()[0].T.Location().String(), "UTC"; got != want {
				t.Errorf("EndLine() t argument time zone: want %v, got %v", want, got)
			}
			if got, want := sb.String(), "encoded content"; got != want {
				t.Errorf("written data: want %q, got %q", want, got)
			}
			if got, want := len(mockedPooler.PutCalls()), 1; got != want {
				t.Errorf("pool Put() call count: want %d, got %d", want, got)
			}
			if got, want := mockedPooler.PutCalls()[0].X, mockedEncoder; got != want {
				t.Errorf("pool Put() argument: want %v, got %v", want, got)
			}
		})
	}
}
