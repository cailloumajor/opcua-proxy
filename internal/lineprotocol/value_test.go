package lineprotocol_test

import (
	"math"
	"reflect"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/lineprotocol"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

func TestNewValueFromVariant(t *testing.T) {
	cases := []struct {
		name        string
		variant     *ua.Variant
		expectValue lineprotocol.Value
		expectError bool
	}{
		{
			name:        "Invalid",
			variant:     ua.MustVariant(nil),
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name:        "Boolean",
			variant:     ua.MustVariant(false),
			expectValue: lineprotocol.MustNewValue(false),
			expectError: false,
		},
		{
			name:        "SByte",
			variant:     ua.MustVariant(int8(1)),
			expectValue: lineprotocol.MustNewValue(int64(1)),
			expectError: false,
		},
		{
			name:        "Byte",
			variant:     ua.MustVariant(uint8(2)),
			expectValue: lineprotocol.MustNewValue(uint64(2)),
			expectError: false,
		},
		{
			name:        "Int16",
			variant:     ua.MustVariant(int16(3)),
			expectValue: lineprotocol.MustNewValue(int64(3)),
			expectError: false,
		},
		{
			name:        "Uint16",
			variant:     ua.MustVariant(uint16(4)),
			expectValue: lineprotocol.MustNewValue(uint64(4)),
			expectError: false,
		},
		{
			name:        "Int32",
			variant:     ua.MustVariant(int32(5)),
			expectValue: lineprotocol.MustNewValue(int64(5)),
			expectError: false,
		},
		{
			name:        "Uint32",
			variant:     ua.MustVariant(uint32(6)),
			expectValue: lineprotocol.MustNewValue(uint64(6)),
			expectError: false,
		},
		{
			name:        "Int64",
			variant:     ua.MustVariant(int64(7)),
			expectValue: lineprotocol.MustNewValue(int64(7)),
			expectError: false,
		},
		{
			name:        "Uint64",
			variant:     ua.MustVariant(uint64(8)),
			expectValue: lineprotocol.MustNewValue(uint64(8)),
			expectError: false,
		},
		{
			name:        "NonFiniteFloat",
			variant:     ua.MustVariant(math.NaN()),
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name:        "Float",
			variant:     ua.MustVariant(float32(9.0)),
			expectValue: lineprotocol.MustNewValue(float64(9.0)),
			expectError: false,
		},
		{
			name:        "Double",
			variant:     ua.MustVariant(float64(10.1)),
			expectValue: lineprotocol.MustNewValue(float64(10.1)),
			expectError: false,
		},
		{
			name:        "InvalidString",
			variant:     ua.MustVariant("aa\xe2"),
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name:        "String",
			variant:     ua.MustVariant("test_string"),
			expectValue: lineprotocol.MustNewValue("test_string"),
			expectError: false,
		},
		{
			name:        "XMLElement",
			variant:     ua.MustVariant(ua.XMLElement("test_xml_element")),
			expectValue: lineprotocol.MustNewValue("test_xml_element"),
			expectError: false,
		},
		{
			name:        "LocalizedText",
			variant:     ua.MustVariant(ua.NewLocalizedText("test_localized_text")),
			expectValue: lineprotocol.MustNewValue("test_localized_text"),
			expectError: false,
		},
		{
			name:        "QualifiedName",
			variant:     ua.MustVariant(&ua.QualifiedName{NamespaceIndex: 42, Name: "test_qualified_name"}),
			expectValue: lineprotocol.MustNewValue("test_qualified_name"),
			expectError: false,
		},
		{
			name:        "InvalidByteString",
			variant:     ua.MustVariant([]byte("aa\xe2")),
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name:        "ByteString",
			variant:     ua.MustVariant([]byte("test_byte_string")),
			expectValue: lineprotocol.MustNewValue([]byte("test_byte_string")),
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := NewValueFromVariant(tc.variant)

			if got, want := v, tc.expectValue; !reflect.DeepEqual(got, want) {
				t.Errorf("field value: want %#v, got %#v", want, got)
			}
			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
		})
	}
}
