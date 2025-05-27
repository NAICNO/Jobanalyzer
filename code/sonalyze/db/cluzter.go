// Data representation of Cluzter data, and parser for CSV files holding those data.
//
// The name "Cluzter" is used for the sinfo-derived cluster data since the name "cluster" was
// already used extensively by the database layer for other things.

package db

import (
	"unsafe"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
)

type CluzterInfo = newfmt.ClusterAttributes

func cluzterInfoSize(c *CluzterInfo) uintptr {
	size := unsafe.Sizeof(*c)
	size += uintptr(len(c.Time))
	size += uintptr(len(c.Cluster))
	for i := range c.Partitions {
		size += cluzterPartitionSize(&c.Partitions[i])
	}
	for i := range c.Nodes {
		size += cluzterNodesSize(&c.Nodes[i])
	}
	return size
}

func cluzterPartitionSize(p *newfmt.ClusterPartition) uintptr {
	size := unsafe.Sizeof(*p)
	size += uintptr(len(p.Name))
	for _, r := range p.Nodes {
		size += stringSize + uintptr(len(r))
	}
	return size
}

func cluzterNodesSize(n *newfmt.ClusterNodes) uintptr {
	size := unsafe.Sizeof(*n)
	for _, name := range n.Names {
		size += stringSize + uintptr(len(name))
	}
	for _, state := range n.States {
		size += stringSize + uintptr(len(state))
	}
	return size
}
