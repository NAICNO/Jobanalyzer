// Data representation of slurm cluster attribute data.
//
// (The name "Cluzter" is used for the sinfo-derived cluster data since the name "cluster" was
// already used extensively by the database layer for other things.)

package repr

import (
	"unsafe"
)

// Simple 2D table of cluster attributes
type CluzterAttributes struct {
	Time    string
	Cluster string
	Slurm   bool
}

func (c *CluzterAttributes) TimeAndNode() (string, string) {
	return c.Time, ""
}

func (c *CluzterAttributes) Size() uintptr {
	size := unsafe.Sizeof(*c)
	size += uintptr(len(c.Time))
	size += uintptr(len(c.Cluster))
	return size
}
