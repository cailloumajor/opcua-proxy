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
		variant     VariantProvider
		expectValue lineprotocol.Value
		expectError bool
	}{
		{
			name: "Unsupported",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDNull },
			},
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name: "Boolean",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDBoolean },
				BoolFunc: func() bool { return false },
			},
			expectValue: lineprotocol.MustNewValue(false),
			expectError: false,
		},
		{
			name: "SByte",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDSByte },
				IntFunc:  func() int64 { return 1 },
			},
			expectValue: lineprotocol.MustNewValue(int64(1)),
			expectError: false,
		},
		{
			name: "Byte",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDByte },
				UintFunc: func() uint64 { return 2 },
			},
			expectValue: lineprotocol.MustNewValue(uint64(2)),
			expectError: false,
		},
		{
			name: "Int16",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDInt16 },
				IntFunc:  func() int64 { return 3 },
			},
			expectValue: lineprotocol.MustNewValue(int64(3)),
			expectError: false,
		},
		{
			name: "Uint16",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDUint16 },
				UintFunc: func() uint64 { return 4 },
			},
			expectValue: lineprotocol.MustNewValue(uint64(4)),
			expectError: false,
		},
		{
			name: "Int32",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDInt32 },
				IntFunc:  func() int64 { return 5 },
			},
			expectValue: lineprotocol.MustNewValue(int64(5)),
			expectError: false,
		},
		{
			name: "Uint32",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDUint32 },
				UintFunc: func() uint64 { return 6 },
			},
			expectValue: lineprotocol.MustNewValue(uint64(6)),
			expectError: false,
		},
		{
			name: "Int64",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDInt64 },
				IntFunc:  func() int64 { return 7 },
			},
			expectValue: lineprotocol.MustNewValue(int64(7)),
			expectError: false,
		},
		{
			name: "Uint64",
			variant: &VariantProviderMock{
				TypeFunc: func() ua.TypeID { return ua.TypeIDUint64 },
				UintFunc: func() uint64 { return 8 },
			},
			expectValue: lineprotocol.MustNewValue(uint64(8)),
			expectError: false,
		},
		{
			name: "NonFiniteFloat",
			variant: &VariantProviderMock{
				TypeFunc:  func() ua.TypeID { return ua.TypeIDFloat },
				FloatFunc: func() float64 { return math.NaN() },
			},
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name: "Float",
			variant: &VariantProviderMock{
				TypeFunc:  func() ua.TypeID { return ua.TypeIDFloat },
				FloatFunc: func() float64 { return 9.0 },
			},
			expectValue: lineprotocol.MustNewValue(float64(9.0)),
			expectError: false,
		},
		{
			name: "Double",
			variant: &VariantProviderMock{
				TypeFunc:  func() ua.TypeID { return ua.TypeIDDouble },
				FloatFunc: func() float64 { return 10.1 },
			},
			expectValue: lineprotocol.MustNewValue(float64(10.1)),
			expectError: false,
		},
		{
			name: "InvalidString",
			variant: &VariantProviderMock{
				TypeFunc:   func() ua.TypeID { return ua.TypeIDString },
				StringFunc: func() string { return "aa\xe2" },
			},
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name: "String",
			variant: &VariantProviderMock{
				TypeFunc:   func() ua.TypeID { return ua.TypeIDString },
				StringFunc: func() string { return "test_string" },
			},
			expectValue: lineprotocol.MustNewValue("test_string"),
			expectError: false,
		},
		{
			name: "XMLElement",
			variant: &VariantProviderMock{
				TypeFunc:   func() ua.TypeID { return ua.TypeIDXMLElement },
				StringFunc: func() string { return "test_xml_element" },
			},
			expectValue: lineprotocol.MustNewValue("test_xml_element"),
			expectError: false,
		},
		{
			name: "LocalizedText",
			variant: &VariantProviderMock{
				TypeFunc:   func() ua.TypeID { return ua.TypeIDLocalizedText },
				StringFunc: func() string { return "test_localized_text" },
			},
			expectValue: lineprotocol.MustNewValue("test_localized_text"),
			expectError: false,
		},
		{
			name: "QualifiedName",
			variant: &VariantProviderMock{
				TypeFunc:   func() ua.TypeID { return ua.TypeIDQualifiedName },
				StringFunc: func() string { return "test_qualified_name" },
			},
			expectValue: lineprotocol.MustNewValue("test_qualified_name"),
			expectError: false,
		},
		{
			name: "InvalidByteString",
			variant: &VariantProviderMock{
				TypeFunc:       func() ua.TypeID { return ua.TypeIDByteString },
				ByteStringFunc: func() []byte { return []byte("aa\xe2") },
			},
			expectValue: lineprotocol.Value{},
			expectError: true,
		},
		{
			name: "ByteString",
			variant: &VariantProviderMock{
				TypeFunc:       func() ua.TypeID { return ua.TypeIDByteString },
				ByteStringFunc: func() []byte { return []byte("test_byte_string") },
			},
			expectValue: lineprotocol.MustNewValue([]byte("test_byte_string")),
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := NewValueFromVariant(tc.variant)

			if msg := testutils.AssertError(t, err, tc.expectError); msg != "" {
				t.Errorf(msg)
			}
			if tc.expectError {
				return
			}
			if got, want := v, tc.expectValue; !reflect.DeepEqual(got, want) {
				t.Errorf("field value: want %#v, got %#v", want, got)
			}
		})
	}
}
