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
