// Data representation of Sonar `sample` data.

package repr

import (
	"go-utils/gpuset"
	"unsafe"

	. "sonalyze/common"
)

const (
	FlagHeartbeat = 1 // Record is a heartbeat record
)

// The core type for process samples is the `Sample`, which represents one log record without the
// per-system data (such as `load`).
//
// After ingestion and initial correction this datum is *strictly* read-only.  It will be accessed
// concurrently without locking from many threads and must not be written by any of them.
//
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
// TODO: OPTIMIZEME: Now that we are caching data we have more opportunities.  Samples are not
// stored individually but can be stored as part of a postprocessed stream keyed by Host, StreamId
// (Pid or JobId), and Cmd.
//
//  - Common fields (maybe Host, Job, Pid, User, Cluster, Cmd) can be lifted out of the structure to
//    a header
//  - Timestamp can be delta-encoded as u16
//  - Version can be removed, as all version-dependent corrections will have been applied during
//    stream postprocessing
//
// It will also be advantageous to store structures in-line in tightly controlled slices rather than
// as individual heap-allocated structures.
//
// Some of these optimizations will complicate the use of the data, obviously.

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

// The LoadDatum represents the `load` field from a record.  The data array is owned by its datum
// and does not share storage with the input.
//
// After ingestion and initial correction this datum is *strictly* read-only.  It will be accessed
// concurrently without locking from many threads and must not be written by any of them.
//
// In older data the data array is represented encoded as that is the most sensible representation
// for the cache.  In newer data the data coming from disk are already decoded and it makes no sense
// to re-encode them here.

type EncodedLoadData struct {
	x any
}

func EncodedLoadDataFromBytes(xs []byte) EncodedLoadData {
	return EncodedLoadData{x: xs}
}

func EncodedLoadDataFromValues(xs []uint64) EncodedLoadData {
	return EncodedLoadData{x: xs}
}

func DecodeEncodedLoadData(e EncodedLoadData) ([]byte, []uint64) {
	switch xs := e.x.(type) {
	case []byte:
		return xs, nil
	case []uint64:
		return nil, xs
	default:
		panic("Unexpected")
	}
}

type LoadDatum struct {
	Timestamp int64
	Hostname  Ustr
	Encoded   EncodedLoadData
}

func (l *LoadDatum) Size() uintptr {
	size := unsafe.Sizeof(*l)
	a, b := DecodeEncodedLoadData(l.Encoded)
	if a != nil {
		size += uintptr(len(a))
	} else {
		size += uintptr(len(b)) * 8
	}
	return size
}

// The same as LoadDatum buf for the "gpuinfo" field that was introduced with Sonar 0.13.

type EncodedGpuData struct {
	x any
}

type PerGpuSample struct {
	FanPct      int // Can go above 100
	PerfMode    int // Typically mode is P<n>, this is <n>
	MemUsedKB   int64
	TempC       int
	PowerDrawW  int
	PowerLimitW int
	CeClockMHz  int
	MemClockMHz int
}

var (
	// MT: Constant after initialization; immutable
	sizeofPerGpuSample uintptr
)

func init() {
	var x PerGpuSample
	sizeofPerGpuSample = unsafe.Sizeof(x)
}

func EncodedGpuDataFromBytes(xs []byte) EncodedGpuData {
	return EncodedGpuData{x: xs}
}

func EncodedGpuDataFromValues(xs []PerGpuSample) EncodedGpuData {
	return EncodedGpuData{x: xs}
}

func DecodeEncodedGpuData(e EncodedGpuData) ([]byte, []PerGpuSample) {
	switch xs := e.x.(type) {
	case []byte:
		return xs, nil
	case []PerGpuSample:
		return nil, xs
	default:
		panic("Unexpected")
	}
}

type GpuDatum struct {
	Timestamp int64
	Hostname  Ustr
	Encoded   EncodedGpuData
}

func (g *GpuDatum) Size() uintptr {
	size := unsafe.Sizeof(*g)
	a, b := DecodeEncodedGpuData(g.Encoded)
	if a != nil {
		size += uintptr(len(a))
	} else {
		size += uintptr(len(b)) * sizeofPerGpuSample
	}
	return size
}
