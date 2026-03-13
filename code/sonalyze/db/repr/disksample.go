package repr

import (
	"time"
	"unsafe"

	. "sonalyze/common"
)

type DiskSample struct {
	Timestamp int64
	Hostname  Ustr
	Name      Ustr
	Major     uint64
	Minor     uint64

	// Following are raw values from /proc/diskstats in the order they appear in that array, do not
	// reorder!  See https://www.kernel.org/doc/Documentation/admin-guide/iostats.rst.
	ReadsCompleted    uint64
	ReadsMerged       uint64
	SectorsRead       uint64
	MsReading         uint64
	WritesCompleted   uint64
	WritesMerged      uint64
	SectorsWritten    uint64
	MsWriting         uint64
	IOsInProgress     uint64
	MsDoingIO         uint64
	WeightedMsDoingIO uint64
	DiscardsCompleted uint64
	DiscardsMerged    uint64
	SectorsDiscarded  uint64
	MsDiscarding      uint64
	FlushesCompleted  uint64
	MsFlushing        uint64
}

func (c *DiskSample) TimeAndNode() (any, string) {
	return time.Unix(c.Timestamp, 0), c.Hostname.String()
}

func (d *DiskSample) Size() uintptr {
	return unsafe.Sizeof(*d)
}
