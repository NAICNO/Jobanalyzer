// Data representation of Sonar `sample` data, and parser for CSV files holding those data.

package db

import (
	"errors"
	"go-utils/gpuset"
	"io"
	"math"
	"slices"
	"time"

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

// The LoadDatum represents the `load` field from a record.  The data array is owned by its datum
// and does not share storage with the input.  It is represented encoded as that is the most
// sensible representation for the cache.
//
// After ingestion and initial correction this datum is *strictly* read-only.  It will be accessed
// concurrently without locking from many threads and must not be written by any of them.

type LoadDatum struct {
	Timestamp int64
	Hostname  Ustr
	Encoded   []byte
}

// The same as LoadDatum buf for the "gpuinfo" field that was introduced with Sonar 0.13.

type GpuDatum struct {
	Timestamp int64
	Hostname  Ustr
	Encoded   []byte
}

// Read a stream of Sonar data records, parse them and return them in order.  Returns the number of
// benign errors, and non-nil error if non-benign error.
//
// Efficiency is a major concern, the parser has been tweaked in many ways to reduce allocation and
// improve parsing speed.

func ParseSonarLog(
	input io.Reader,
	ustrs UstrAllocator,
	verbose bool,
) (
	samples []*Sample,
	loadData []*LoadDatum,
	gpuData []*GpuDatum,
	softErrors int,
	err error,
) {
	const (
		unknownFormat = iota
		untaggedFormat
		taggedFormat
	)

	samples = make([]*Sample, 0)
	loadData = make([]*LoadDatum, 0)
	gpuData = make([]*GpuDatum, 0)
	tokenizer := NewTokenizer(input)
	v060 := ustrs.Alloc("0.6.0")
	heartbeat := ustrs.Alloc("_heartbeat_")
	endOfInput := false

LineLoop:
	for !endOfInput {
		// Find the fields and then convert them.  Duplicates are not allowed.  Mandatory fields are
		// really required.  The sentinels are not zero because zeroes are valid values from the
		// input.  Keep the sentinels in sync with the code below that inserts default values after
		// parsing!
		var (
			version                  = UstrEmpty
			timestamp        int64   = math.MaxInt64
			hostname                 = UstrEmpty
			numCores         uint32  = math.MaxUint32
			memTotalKB       uint64  = math.MaxUint64
			user                     = UstrEmpty
			pid              uint32  = math.MaxUint32
			ppid             uint32  = math.MaxUint32
			jobId            uint32  = math.MaxUint32
			command                  = UstrEmpty
			cpuPct           float32 = math.MaxFloat32
			cpuKB            uint64  = math.MaxUint64
			rssAnonKB        uint64  = math.MaxUint64
			gpus                     = gpuset.EmptyGpuSet()
			haveGpus                 = false
			gpuPct           float32 = math.MaxFloat32
			gpuMemPct        float32 = math.MaxFloat32
			gpuKB            uint64  = math.MaxUint64
			gpuFail          uint8   = math.MaxUint8
			cpuTimeSec       uint64  = math.MaxUint64
			rolledup         uint32  = math.MaxUint32
			load             []byte
			gpuinfo          []byte
			format           = unknownFormat
			untaggedPosition = 0
		)

	FieldLoop:
		for {
			var start, lim, eqloc int
			var matched bool
			start, lim, eqloc, err = tokenizer.Get()
			if err != nil {
				if !errors.Is(err, SyntaxErr) {
					return
				}
				tokenizer.ScanEol()
				softErrors++
				continue LineLoop
			}

			if start == CsvEol {
				break FieldLoop
			}

			if start == CsvEof {
				endOfInput = true
				break FieldLoop
			}

			if format == unknownFormat {
				if eqloc == CsvEqSentinel {
					format = untaggedFormat
					version = v060
				} else {
					format = taggedFormat
				}
			}

			// Regarding timestamps: This is the format we use in the Sonar logs, but the nano part
			// is often omitted by our formatters:
			//
			//  "2006-01-02T15:04:05.999999999-07:00"
			//
			// time.RFC3339Nano handles +/- for the tz offset and also will allow the nano part to
			// be missing.

			switch format {
			case unknownFormat:
				panic("Unexpected case")

			case untaggedFormat:
				// Old old format (current on Saga and Fram as of 2024-03-04)
				// 0  timestamp
				// 1  hostname
				// 2  numcores
				// 3  username
				// 4  jobid
				// 5  command
				// 6  cpu_pct
				// 7  mem_kib
				//
				// New old format (what was briefly deployed on the UiO ML nodes)
				// 8  gpus bitvector
				// 9  gpu_pct
				// 10 gpumem_pct
				// 11 gpumem_kib
				//
				// Newer old format (again briefly used on the UiO ML nodes)
				// 12 cputime_sec
				val := tokenizer.BufSlice(start, lim)
				switch untaggedPosition {
				case 0:
					var tmp time.Time
					tmp, err = time.Parse(time.RFC3339Nano, string(val))
					if err != nil {
						softErrors++
						tokenizer.ScanEol()
						continue LineLoop
					}
					timestamp = tmp.Unix()
				case 1:
					hostname = ustrs.AllocBytes(val)
				case 2:
					var tmp uint64
					tmp, err = parseUint(val)
					if err != nil {
						softErrors++
						tokenizer.ScanEol()
						continue LineLoop
					}
					numCores = uint32(tmp)
				case 3:
					user = ustrs.AllocBytes(val)
				case 4:
					var tmp uint64
					tmp, err = parseUint(val)
					if err != nil {
						softErrors++
						tokenizer.ScanEol()
						continue LineLoop
					}
					jobId = uint32(tmp)
					pid = jobId
				case 5:
					command = ustrs.AllocBytes(val)
				case 6:
					var tmp float64
					tmp, err = parseFloat(val, true)
					if err != nil {
						softErrors++
						tokenizer.ScanEol()
						continue LineLoop
					}
					cpuPct = float32(tmp)
				case 7:
					var tmp uint64
					tmp, err = parseUint(val)
					if err != nil {
						softErrors++
						tokenizer.ScanEol()
						continue LineLoop
					}
					cpuKB = tmp
				default:
					// Ignore any remaining fields - they are not in most untagged data.
				}
				untaggedPosition++
				matched = true

			case taggedFormat:
				// NOTE, in error cases below we don't extract the offending field b/c it seems the
				// optimizer will hoist the (technically effect-free) extraction out of the parsing
				// switch and slow everything down tremendously.

				if eqloc == CsvEqSentinel {
					// Invalid field syntax: Drop the field but keep the record
					if verbose {
						Log.Infof(
							"Dropping field with bad form: %s",
							"(elided)", /*tokenizer.BufSubstringSlow(start, lim), - see NOTE above*/
						)
					}
					softErrors++
					continue FieldLoop
				}

				// No need to check that BufAt(start+1) is valid: The first two characters will
				// always be present because eqloc is either CsvEqSentinel (handled above) or
				// greater than zero (the field name is never empty).
				switch tokenizer.BufAt(start) {
				case 'c':
					switch tokenizer.BufAt(start + 1) {
					case 'o':
						if val, ok := match(tokenizer, start, lim, eqloc, "cores"); ok {
							var tmp uint64
							tmp, err = parseUint(val)
							numCores = uint32(tmp)
							matched = true
						}
					case 'm':
						if val, ok := match(tokenizer, start, lim, eqloc, "cmd"); ok {
							command = ustrs.AllocBytes(val)
							matched = true
						}
					case 'p':
						if lim-start >= 4 {
							switch tokenizer.BufAt(start + 3) {
							case '%':
								if val, ok := match(tokenizer, start, lim, eqloc, "cpu%"); ok {
									var tmp float64
									tmp, err = parseFloat(val, true)
									cpuPct = float32(tmp)
									matched = true
								}
							case 'k':
								if val, ok := match(tokenizer, start, lim, eqloc, "cpukib"); ok {
									cpuKB, err = parseUint(val)
									matched = true
								}
							case 't':
								if val, ok := match(tokenizer, start, lim, eqloc, "cputime_sec"); ok {
									cpuTimeSec, err = parseUint(val)
									matched = true
								}
							}
						}
					}
				case 'g':
					if lim-start >= 4 {
						switch tokenizer.BufAt(start + 3) {
						case '%':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpu%"); ok {
								var tmp float64
								tmp, err = parseFloat(val, true)
								gpuPct = float32(tmp)
								matched = true
							}
						case 'f':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpufail"); ok {
								var tmp uint64
								tmp, err = parseUint(val)
								gpuFail = uint8(tmp)
								matched = true
							}
						case 'i':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpuinfo"); ok {
								gpuinfo = slices.Clone(val)
								matched = true
							}
						case 'k':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpukib"); ok {
								gpuKB, err = parseUint(val)
								matched = true
							}
						case 'm':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpumem%"); ok {
								var tmp float64
								tmp, err = parseFloat(val, true)
								gpuMemPct = float32(tmp)
								matched = true
							}
						case 's':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpus"); ok {
								gpus, err = gpuset.NewGpuSet(string(val))
								haveGpus = true
								matched = true
							}
						}
					}
				case 'h':
					if val, ok := match(tokenizer, start, lim, eqloc, "host"); ok {
						hostname = ustrs.AllocBytes(val)
						matched = true
					}
				case 'j':
					if val, ok := match(tokenizer, start, lim, eqloc, "job"); ok {
						var tmp uint64
						tmp, err = parseUint(val)
						jobId = uint32(tmp)
						matched = true
					}
				case 'l':
					if val, ok := match(tokenizer, start, lim, eqloc, "load"); ok {
						load = slices.Clone(val)
						matched = true
					}
				case 'm':
					if val, ok := match(tokenizer, start, lim, eqloc, "memtotalkib"); ok {
						memTotalKB, err = parseUint(val)
						matched = true
					}
				case 'p':
					if val, ok := match(tokenizer, start, lim, eqloc, "pid"); ok {
						var tmp uint64
						tmp, err = parseUint(val)
						pid = uint32(tmp)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "ppid"); ok {
						var tmp uint64
						tmp, err = parseUint(val)
						ppid = uint32(tmp)
						matched = true
					}
				case 'r':
					if val, ok := match(tokenizer, start, lim, eqloc, "rssanonkib"); ok {
						rssAnonKB, err = parseUint(val)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "rolledup"); ok {
						var tmp uint64
						tmp, err = parseUint(val)
						rolledup = uint32(tmp)
						matched = true
					}
				case 't':
					if val, ok := match(tokenizer, start, lim, eqloc, "time"); ok {
						var tmp time.Time
						tmp, err = time.Parse(time.RFC3339Nano, string(val))
						timestamp = tmp.Unix()
						matched = true
					}
				case 'u':
					if val, ok := match(tokenizer, start, lim, eqloc, "user"); ok {
						user = ustrs.AllocBytes(val)
						matched = true
					}
				case 'v':
					if val, ok := match(tokenizer, start, lim, eqloc, "v"); ok {
						version = ustrs.AllocBytes(val)
						matched = true
					}
				}
				// Four cases:
				//
				//   matched && !failed - field matched a tag, value is good
				//   matched && failed - field matched a tag, value is bad
				//   !matched && !failed - field did not match any tag
				//   !matched && failed - impossible
				//
				// The second case suggests something bad, so discard the record in this case.  Note
				// this is actually the same as just `failed` due to the fourth case.
				if !matched {
					if verbose {
						Log.Warningf(
							"Dropping field with unknown name: %s",
							"(elided)", /* tokenizer.BufSubstringSlow(start, eqloc-1), -
							   see NOTE above */
						)
					}
					if err == nil {
						softErrors++
					}
				}
				if err != nil {
					if verbose {
						Log.Warningf(
							"Dropping record with illegal/unparseable value: %s %v",
							"(elided)", /*tokenizer.BufSubstringSlow(start, lim), - see NOTE above */
							err,
						)
					}
					softErrors++
					tokenizer.ScanEol()
					continue LineLoop
				}

			default:
				panic("Unexpected case")
			}
		} // end FieldLoop

		// Skip entirely empty records.
		if format == unknownFormat {
			continue LineLoop
		}

		// Untagged records do not have optional trailing fields.
		if format == untaggedFormat && untaggedPosition < 8 {
			if verbose {
				Log.Infof(
					"Dropping untagged record with missing fields, got only %d fields",
					untaggedPosition,
				)
			}
			softErrors++
			continue LineLoop
		}

		// Fields have been parsed, now check them
		irritants := ""
		if version == UstrEmpty || timestamp == math.MaxInt64 || hostname == UstrEmpty ||
			command == UstrEmpty {
			if version == UstrEmpty {
				irritants += "version "
			}
			if timestamp == math.MaxInt64 {
				irritants += "time "
			}
			if hostname == UstrEmpty {
				irritants += "host "
			}
			if command == UstrEmpty {
				irritants += "cmd "
			}
		}
		if command != heartbeat && user == UstrEmpty {
			irritants += "user "
		}
		if irritants != "" {
			if verbose {
				Log.Warningf("Dropping record with missing mandatory field(s): %s", irritants)
			}
			softErrors++
			continue LineLoop
		}

		// Fill in default data for optional fields.  Keep this code in sync with initialization
		// above!
		if numCores == math.MaxUint32 {
			numCores = 0
		}
		if memTotalKB == math.MaxUint64 {
			memTotalKB = 0
		}
		if jobId == math.MaxUint32 {
			jobId = 0
		}
		if pid == math.MaxUint32 {
			pid = 0
		}
		if ppid == math.MaxUint32 {
			ppid = 0
		}
		if cpuPct == math.MaxFloat32 {
			cpuPct = 0
		}
		if cpuKB == math.MaxUint64 {
			cpuKB = 0
		}
		if rssAnonKB == math.MaxUint64 {
			rssAnonKB = 0
		}
		if !haveGpus {
			gpus = gpuset.EmptyGpuSet()
		}
		if gpuPct == math.MaxFloat32 {
			gpuPct = 0
		}
		if gpuMemPct == math.MaxFloat32 {
			gpuMemPct = 0
		}
		if gpuKB == math.MaxUint64 {
			gpuKB = 0
		}
		if gpuFail == math.MaxUint8 {
			gpuFail = 0
		}
		if cpuTimeSec == math.MaxUint64 {
			cpuTimeSec = 0
		}
		if rolledup == math.MaxUint32 {
			rolledup = 0
		}

		flags := uint8(0)
		if command == heartbeat {
			flags |= FlagHeartbeat
		}
		samples = append(samples, &Sample{
			Version:    version,
			Timestamp:  timestamp,
			Hostname:   hostname,
			Cores:      numCores,
			MemtotalKB: memTotalKB,
			User:       user,
			Pid:        pid,
			Ppid:       ppid,
			Job:        jobId,
			Cmd:        command,
			CpuPct:     cpuPct,
			CpuKB:      cpuKB,
			RssAnonKB:  rssAnonKB,
			Gpus:       gpus,
			GpuPct:     gpuPct,
			GpuMemPct:  gpuMemPct,
			GpuKB:      gpuKB,
			GpuFail:    gpuFail,
			CpuTimeSec: cpuTimeSec,
			Rolledup:   rolledup,
			Flags:      flags,
		})
		if load != nil {
			loadData = append(loadData, &LoadDatum{
				Timestamp: timestamp,
				Hostname:  hostname,
				Encoded:   load,
			})
		}
		if gpuinfo != nil {
			gpuData = append(gpuData, &GpuDatum{
				Timestamp: timestamp,
				Hostname:  hostname,
				Encoded:   gpuinfo,
			})
		}
	}

	err = nil
	return
}
