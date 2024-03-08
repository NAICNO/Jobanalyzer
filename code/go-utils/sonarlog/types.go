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
//    GpuFail is a single bit at present.
//
// It seems likely that if we applied all of these we could save another 30 bytes easily.

type SonarReading struct {
	Timestamp   int64 // seconds since year 1
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
	Gpus        Ustr
	GpuPct      float32
	GpuMemPct   float32
	Rolledup    uint32
	GpuFail     uint8
}

type SonarHeartbeat struct {
	Timestamp int64 // seconds since year 1
	Version   Ustr
	Cluster   Ustr
	Host      Ustr
}
