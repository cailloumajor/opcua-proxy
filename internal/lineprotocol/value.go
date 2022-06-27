package lineprotocol

import (
	"fmt"

	"github.com/gopcua/opcua/ua"
	lp "github.com/influxdata/line-protocol/v2/lineprotocol"
)

// NewValueFromVariant creates a line protocol field value from an OPC-UA variant.
func NewValueFromVariant(variant *ua.Variant) (lp.Value, error) {
	var v lp.Value
	var err error

	t := variant.Type()

	switch t {
	case ua.TypeIDBoolean:
		v = lp.BoolValue(variant.Bool())
	case ua.TypeIDSByte, ua.TypeIDInt16, ua.TypeIDInt32, ua.TypeIDInt64:
		v = lp.IntValue(variant.Int())
	case ua.TypeIDByte, ua.TypeIDUint16, ua.TypeIDUint32, ua.TypeIDUint64:
		v = lp.UintValue(variant.Uint())
	case ua.TypeIDFloat, ua.TypeIDDouble:
		var ok bool
		v, ok = lp.FloatValue(variant.Float())
		if !ok {
			err = fmt.Errorf("unsupported non-finite float: %v", variant.Float())
		}
	case ua.TypeIDString, ua.TypeIDXMLElement, ua.TypeIDLocalizedText, ua.TypeIDQualifiedName:
		var ok bool
		v, ok = lp.StringValue(variant.String())
		if !ok {
			err = fmt.Errorf("invalid string")
		}
	case ua.TypeIDByteString:
		var ok bool
		v, ok = lp.StringValueFromBytes(variant.ByteString())
		if !ok {
			err = fmt.Errorf("invalid byte string")
		}
	default:
		err = fmt.Errorf("unsupported variant type: %q", t)
	}

	return v, err
}
