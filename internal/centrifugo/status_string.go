// Code generated by "stringer -linecomment -trimprefix status -type status"; DO NOT EDIT.

package centrifugo

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[statusOK-0]
	_ = x[statusOPCUaNotConnected-1]
}

const _status_name = "Everything OKOPC-UA not connected"

var _status_index = [...]uint8{0, 13, 33}

func (i status) String() string {
	if i >= status(len(_status_index)-1) {
		return "status(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _status_name[_status_index[i]:_status_index[i+1]]
}
