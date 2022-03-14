package opcua

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/gopcua/opcua/ua"
)

// Node wraps an OPC-UA NodeID.
type Node struct {
	t         ua.IDType
	numericID uint32
	stringID  string
}

//UnmarshalJSON implements json.Unmarshaler.
func (n *Node) UnmarshalJSON(b []byte) error {
	var u interface{}
	if err := json.Unmarshal(b, &u); err != nil {
		return err
	}

	switch v := u.(type) {
	case string:
		n.t = ua.IDTypeString
		n.stringID = v
	case float64:
		i, f := math.Modf(v)
		if f != 0 {
			return fmt.Errorf("not a whole number: %v", v)
		}
		p, err := strconv.ParseUint(fmt.Sprintf("%.f", i), 10, 32)
		if err != nil {
			return fmt.Errorf("integer number error: %w", err)
		}
		n.t = ua.IDTypeNumeric
		n.numericID = uint32(p)
	default:
		return fmt.Errorf("unsupported type: %T", u)
	}

	return nil
}

// NodeID returns the wrapped NodeID.
func (n *Node) NodeID(ns uint16) *ua.NodeID {
	switch n.t {
	case ua.IDTypeNumeric:
		return ua.NewNumericNodeID(ns, n.numericID)
	case ua.IDTypeString:
		return ua.NewStringNodeID(ns, n.stringID)
	default:
		return &ua.NodeID{}
	}
}
