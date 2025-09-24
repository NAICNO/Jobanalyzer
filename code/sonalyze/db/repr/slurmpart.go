// Data representation of slurm cluster partition data.
//
// (The name "Cluzter" is used for the sinfo-derived cluster data since the name "cluster" was
// already used extensively by the database layer for other things.)

package repr

import (
	"unsafe"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
)

// A 3D volume where each partition has a name and a node range
type CluzterPartitions struct {
	Time       string
	Cluster    string
	Partitions []newfmt.ClusterPartition
}

func (c *CluzterPartitions) TimeAndNode() (any, string) {
	return c.Time, ""
}

func (c *CluzterPartitions) Size() uintptr {
	size := unsafe.Sizeof(*c)
	size += uintptr(len(c.Time))
	size += uintptr(len(c.Cluster))
	size += unsafe.Sizeof(c.Partitions[0]) * uintptr(cap(c.Partitions))
	for i := range c.Partitions {
		p := &c.Partitions[i]
		size += uintptr(len(p.Name))
		size += unsafe.Sizeof(p.Nodes[0]) * uintptr(cap(p.Nodes))
		for _, nr := range p.Nodes {
			size += uintptr(len(nr))
		}
	}
	return size
}
