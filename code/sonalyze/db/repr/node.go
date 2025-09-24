// Data representation of sysinfo 'node' data

package repr

import (
	"unsafe"
)

// SysinfoNodeData is basically a view on newfmt.SysinfoAttributes where
// newfmt=github.com/NordicHPC/sonar/util/formats/newfmt.  The reason it is a separate view is that
// that data structure carries some fields that should not be visible here.
type SysinfoNodeData struct {
	Time           string
	Cluster        string
	Node           string
	OsName         string
	OsRelease      string
	Architecture   string
	Sockets        uint64
	CoresPerSocket uint64
	ThreadsPerCore uint64
	CpuModel       string
	Memory         uint64
	TopoSVG        string
	Distances      [][]uint64
}

func (n *SysinfoNodeData) TimeAndNode() (any, string) {
	return n.Time, n.Node
}

func (d *SysinfoNodeData) Size() uintptr {
	size := unsafe.Sizeof(*d)
	size += uintptr(len(d.Time))
	size += uintptr(len(d.Cluster))
	size += uintptr(len(d.Node))
	size += uintptr(len(d.OsName))
	size += uintptr(len(d.OsRelease))
	size += uintptr(len(d.Architecture))
	size += uintptr(len(d.CpuModel))
	size += uintptr(len(d.TopoSVG))
	size += unsafe.Sizeof(d.Distances[0][0]) * uintptr(len(d.Distances)*len(d.Distances[0]))
	return size
}
