// Data representation of Sonar `sample` data, ie, per-process sample values.

package repr

import (
	"go-utils/gpuset"
	"unsafe"

	. "sonalyze/common"
)

const (
	FlagHeartbeat = 1 // Record is a heartbeat record
)

// The core type for process samples is the `Sample`, which represents one log record for one process.
//
// After ingestion and initial correction this datum is *strictly* read-only.  It will be accessed
// concurrently without locking from many threads and must not be written by any of them.
//
// This is effectively a flattened view of newfmt.SampleProcess where newfmt is
// github.com/NordicHPC/sonar/util/formats/newfmt.  Fields from its parent types newfmt.SampleJob
// and newfmt.SampleAttributes have been moved into it.
//
// The reason it is a separate view is that that data structure carries some fields that should not
// be visible here, but in addition, we want this structure to be pointer-free and as small as
// possible, as there are very many of them in memory at the same time.  Also, historically these
// data were all part of the same record.
//
// Memory use.
//
// A huge number of these (about 10e6 records per month for Saga, probably 3-4x that for Betzy) may
// be in memory at the same time when processing logs, and there may additionally be other records
// loaded and cached for concurrent queries, so several optimizations have been applied:
//
//  - all fields are pointer free, so these structures don't need to be traced by the GC
//  - strings are hash-consed into Ustr, which takes 4 bytes
//  - fields in the structure have been ordered largest-to-smallest in order to pack the structure,
//    the Go compiler does not do this
//  - the timestamp has been reduced from a 24-byte structure with a pointer to an 8-byte second
//    value, we lose tz info but we never used that anyway and always assumed UTC
//
// Optimizations so far have brought the size from 240 bytes to 104 bytes.
//
// TODO: OPTIMIZEME: Further optimizations are possible:
//
//  - Timestamp can be reduced to uint32
//  - CpuPct, GpuMemPct, GpuPct can be float16 or 16-bit fixedpoint or simply uint16, the value
//    scaled by 10 (ie integer per mille - change the field names)
//  - There are many fields that have unused bits, for example, Ustr is unlikely ever to need more
//    than 24 bits, most memory sizes need more than 32 bits (4GB) but maybe not more than 40 (1TB),
//    Job and Process IDs are probably 24 bits or so, and Rolledup is unlikely to be more than 16
//    bits.  GpuFail and Flags are single bits at present.
//  - Indeed, MemtotalKB and Cores are considered obsolete and could just be removed - will only
//    affect the output of `parse`.
//
// It seems likely that if we applied all of these we could save another 30 bytes.
//
// TODO: OPTIMIZEME: Now that we are caching data we have more opportunities, and the new external
// representations already encode these:
//
//  - Common fields (maybe Host, Job, Pid, User, Cluster, Cmd) can be lifted out of the structure to
//    a header
//
// Additionally:
//
//  - Timestamp can be delta-encoded as u16
//  - Version can be removed, as all version-dependent corrections will have been applied during
//    stream postprocessing
//
// It will also be advantageous to store structures in-line in tightly controlled slices rather than
// as individual heap-allocated structures.
//
// Some of these optimizations will complicate the use of the data, obviously.  Also, with the
// future belonging to a proper database, we should await the needs resulting from that
// reengineering.

type Sample struct {
	Timestamp  int64
	MemtotalKB uint64
	CpuKB      uint64
	RssAnonKB  uint64
	GpuKB      uint64
	CpuTimeSec uint64
	Version    Ustr
	Cluster    Ustr
	Hostname   Ustr
	Cores      uint32
	User       Ustr
	Job        uint32
	Pid        uint32
	Ppid       uint32
	Cmd        Ustr
	CpuPct     float32
	Gpus       gpuset.GpuSet
	GpuPct     float32
	GpuMemPct  float32
	Rolledup   uint32
	GpuFail    uint8
	Flags      uint8
}

var (
	// MT: Constant after initialization; immutable
	SizeofSample uintptr
)

func init() {
	var x Sample
	SizeofSample = unsafe.Sizeof(x)
}
