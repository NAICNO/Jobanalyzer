// Data representation of slurm cluster node data.
//
// (The name "Cluzter" is used for the sinfo-derived cluster data since the name "cluster" was
// already used extensively by the database layer for other things.)

package repr

import (
	"unsafe"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
)

// A 3D volume where each node entry has a node range and a state set
type CluzterNodes struct {
	Time    string
	Cluster string
	Nodes   []newfmt.ClusterNodes
}

func (c *CluzterNodes) TimeAndNode() (any, string) {
	return c.Time, ""
}

func (c *CluzterNodes) Size() uintptr {
	size := unsafe.Sizeof(*c)
	size += uintptr(len(c.Time))
	size += uintptr(len(c.Cluster))
	size += unsafe.Sizeof(c.Nodes[0]) * uintptr(cap(c.Nodes))
	for i := range c.Nodes {
		p := &c.Nodes[i]
		size += unsafe.Sizeof(p.Names[0]) * uintptr(cap(p.Names))
		for _, name := range p.Names {
			size += uintptr(len(name))
		}
		size += StringSize * uintptr(cap(p.States))
		for _, state := range p.States {
			size += uintptr(len(state))
		}
	}
	return size
}
