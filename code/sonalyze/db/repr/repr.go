package repr

import (
	"unsafe"
)

var (
	// MT: Constant after initialization; immutable
	PointerSize uintptr
	StringSize  uintptr
)

func init() {
	var x *int
	PointerSize = unsafe.Sizeof(x)
	var s string
	StringSize = unsafe.Sizeof(s)
}

// `Filterable` is implemented by data representations that can be filtered by the standard
// ApplyFilter function in data/common/filter.go.  The timeVal must be either an RFC3339 timestamp
// string or a time.Time value.  If a string, it will be parsed into a time value.

type Filterable interface {
	TimeAndNode() (timeVal any, nodeName string)
}
