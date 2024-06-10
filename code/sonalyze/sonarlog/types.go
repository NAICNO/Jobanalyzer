// Core types for Sonar `ps` log data.

package sonarlog

import (
	"go-utils/gpuset"
	. "sonalyze/common"
)

const (
	FlagHeartbeat = 1 // Record is a heartbeat record
)

// The core type is the `Sample`, which represents one log record.
//
// A huge number of these (about 10e6 records per month for Saga, probably 4x that for Betzy) may be
// in memory at the same time when processing logs, so several optimizations have been applied:
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
// Further optimizations are possible:
//
//  - Timestamp can be reduced to uint32
//  - CpuPct, GpuMemPct, GpuPct can be float16 or 16-bit fixedpoint or simply uint16, the value
//    scaled by 10 (ie integer per mille - change the field names)
//  - There are many fields that have unused bits, for example, Ustr is unlikely ever to need more
//    than 24 bits, most memory sizes need more than 32 bits (4GB) but maybe not more than 40 (1TB),
//    Job and Process IDs are probably 24 bits or so, and Rolledup is unlikely to be more than 16
//    bits.  GpuFail and Flags are single bits at present.
//  - Indeed, MemtotalKib and Cores are considered obsolete and could just be removed - will only
//    affect the output of `parse`.
//
// It seems likely that if we applied all of these we could save another 30 bytes.
//
// Looking further, once we start caching data we will have more opportunities.  In that case,
// samples will not be stored individually but as part of a postprocessed stream keyed by Host,
// StreamId (Pid or JobId), and Cmd.
//
//  - Common fields (maybe Host, Job, Pid, User, Cluster, Cmd) can be lifted out of the structure to
//    a header
//  - Timestamp can be delta-encoded as u16
//  - Version can be removed, as all version-dependent corrections will have been applied during
//    stream postprocessing
//
// When caching, it will also be advantageous to store structures in-line in tightly controlled
// slices rather than as individual heap-allocated structures.
//
// Some of these optimizations will complicate the use of the data, obviously.

type Sample struct {
	Timestamp   int64
	MemtotalKib uint64
	CpuKib      uint64
	RssAnonKib  uint64
	GpuKib      uint64
	CpuTimeSec  uint64
	Version     Ustr
	Cluster     Ustr
	Host        Ustr
	Cores       uint32
	User        Ustr
	Job         uint32
	Pid         uint32
	Cmd         Ustr
	CpuPct      float32
	Gpus        gpuset.GpuSet
	GpuPct      float32
	CpuUtilPct  float32
	GpuMemPct   float32
	Rolledup    uint32
	GpuFail     uint8
	Flags       uint8
}

// A sample stream is just a list of samples.

type SampleStream []*Sample

// A bag of streams.  The constraints on the individual streams in terms of uniqueness and so on
// depends on how they were merged and are not implied by the type.

type SampleStreams []*SampleStream

// Earliest and latest time stamps found in a set of records.

type Timebound struct {
	Earliest int64
	Latest   int64
}

// Map from host name to bounds for the host name

type Timebounds map[Ustr]Timebound
