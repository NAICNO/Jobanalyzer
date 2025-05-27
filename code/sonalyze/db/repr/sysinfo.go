package repr

import (
	"unsafe"

	"go-utils/config"
)

type SysinfoData config.NodeConfigRecord

func SysinfoDataSize(d *SysinfoData) uintptr {
	size := unsafe.Sizeof(*d)
	size += uintptr(len(d.Timestamp))
	size += uintptr(len(d.Hostname))
	size += uintptr(len(d.Description))
	// Ignore Metadata
	return size
}

