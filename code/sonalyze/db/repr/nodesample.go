// Data representation of sample 'node' data

package repr

import (
	"time"
	"unsafe"

	. "sonalyze/common"
)

// SampleNodeData is basically a view on newfmt.SampleAttributes.SampleSystem where
// newfmt=github.com/NordicHPC/sonar/util/formats/newfmt.
type NodeSample struct {
	Timestamp        int64
	Hostname         Ustr
	UsedMemory       uint64
	Load1            float64
	Load5            float64
	Load15           float64
	RunnableEntities uint64
	ExistingEntities uint64
}

func (c *NodeSample) TimeAndNode() (any, string) {
	return time.Unix(c.Timestamp, 0), c.Hostname.String()
}

func (d *NodeSample) Size() uintptr {
	return unsafe.Sizeof(*d)
}
