// Data representation of sysinfo 'Card' data.

package repr

import (
	"unsafe"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
)

type SysinfoCardData struct {
	Time string
	Node string
	*newfmt.SysinfoGpuCard
}

func (c *SysinfoCardData) TimeAndNode() (string, string) {
	return c.Time, c.Node
}

func (d *SysinfoCardData) Size() uintptr {
	size := unsafe.Sizeof(*d)
	size += uintptr(len(d.Time))
	size += uintptr(len(d.Node))
	s := d.SysinfoGpuCard
	size += unsafe.Sizeof(*s)
	size += uintptr(len(s.UUID))
	size += uintptr(len(s.Address))
	size += uintptr(len(s.Manufacturer))
	size += uintptr(len(s.Model))
	size += uintptr(len(s.Architecture))
	size += uintptr(len(s.Driver))
	size += uintptr(len(s.Firmware))
	return size
}
