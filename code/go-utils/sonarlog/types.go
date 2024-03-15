package sonarlog

// Memory use.
//
// A huge number of these (about 10e6 records per month for Saga) may be in memory at the same time
// when processing logs, so several optimizations have been applied:
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
//  - CpuPct, GpuMemPct, GpuPct can be float16 or 16-bit fixedpoint.
//  - There are many fields that have unused bits, for example, Ustr is unlikely ever to need
//    more than 24 bits, most memory sizes need more than 32 bits but not more than 38, Job and
//    Process IDs are probably 24 bits or so, and Rolledup is unlikely to be more than 16 bits.
//    GpuFail and Flags are single bits at present.
//
// It seems likely that if we applied all of these we could save another 30 bytes easily.

const (
	FlagHeartbeat = 1 // Record is a heartbeat record
)

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
	Gpus        GpuSet
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
