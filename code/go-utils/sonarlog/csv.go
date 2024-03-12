package sonarlog

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

// Sonar data intermingle readings and heartbeats (though really only in the tagged format).  Read a
// stream of records, parse them into separate buckets and return the buckets.  Returns the number
// of benign errors, and non-nil error if non-benign error.  The records in the buckets are in the
// order they appear in the input.
//
// Note wrt parsing floats: According to documentation, strconv.ParseFloat() accepts nan, inf, +inf,
// -inf, infinity, +infinity and -infinity, case-insensitively.  Based on experimentation, the rust
// to_string() formatter will produce "NaN", "inf" and "-inf", with that capitalization (weird).  So
// ingesting CSV data from Rust should not be a problem.

func ParseSonarLog(
	input io.Reader,
	ustrs UstrAllocator,
) (
	readings []*SonarReading,
	heartbeats []*SonarHeartbeat,
	discarded int,
	err error,
) {
	const (
		unknownFormat = iota
		untaggedFormat
		taggedFormat
	)

	readings = make([]*SonarReading, 0)
	heartbeats = make([]*SonarHeartbeat, 0)
	tokenizer := NewTokenizer(input)
	v060 := ustrs.Alloc("0.6.0")
	heartbeat := ustrs.Alloc("_heartbeat_")
	endOfInput := false

LineLoop:
	for !endOfInput {
		// Find the fields and then convert them.  Duplicates are not allowed.  Mandatory
		// fields are really required.  The sentinels are not zero because zero are valid values
		// from the input.  Keep the sentinels in sync with code below that inserts default values
		// after parsing!
		var (
			version                  = UstrEmpty
			timestamp        int64   = math.MaxInt64
			hostname                 = UstrEmpty
			numCores         uint32  = math.MaxUint32
			memTotalKib      uint64  = math.MaxUint64
			user                     = UstrEmpty
			pid              uint32  = math.MaxUint32
			jobId            uint32  = math.MaxUint32
			command                  = UstrEmpty
			cpuPct           float32 = math.MaxFloat32
			cpuKib           uint64  = math.MaxUint64
			rssAnonKib       uint64  = math.MaxUint64
			gpus                     = EmptyGpuSet()
			haveGpus                 = false
			gpuPct           float32 = math.MaxFloat32
			gpuMemPct        float32 = math.MaxFloat32
			gpuKib           uint64  = math.MaxUint64
			gpuFail          uint8   = math.MaxUint8
			cpuTimeSec       uint64  = math.MaxUint64
			rolledup         uint32  = math.MaxUint32
			format                   = unknownFormat
			untaggedPosition         = 0
			anyFields                = false
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
				discarded++
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
				panic("Unexpected state - unknown format")

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
				val := tokenizer.BufSubstring(start, lim)
				switch untaggedPosition {
				case 0:
					var tmp time.Time
					tmp, err = time.Parse(time.RFC3339Nano, val)
					if err != nil {
						discarded++
						tokenizer.ScanEol()
						continue LineLoop
					}
					timestamp = tmp.Unix()
				case 1:
					hostname = ustrs.Alloc(val)
				case 2:
					var tmp uint64
					tmp, err = strconv.ParseUint(val, 10, 64)
					if err != nil {
						discarded++
						tokenizer.ScanEol()
						continue LineLoop
					}
					numCores = uint32(tmp)
				case 3:
					user = ustrs.Alloc(val)
				case 4:
					var tmp uint64
					tmp, err = strconv.ParseUint(val, 10, 64)
					if err != nil {
						discarded++
						tokenizer.ScanEol()
						continue LineLoop
					}
					jobId = uint32(tmp)
					pid = jobId
				case 5:
					command = ustrs.Alloc(val)
				case 6:
					var tmp float64
					tmp, err = strconv.ParseFloat(val, 64)
					if err != nil {
						discarded++
						tokenizer.ScanEol()
						continue LineLoop
					}
					cpuPct = float32(tmp)
				case 7:
					var tmp uint64
					tmp, err = strconv.ParseUint(val, 10, 64)
					if err != nil {
						discarded++
						tokenizer.ScanEol()
						continue LineLoop
					}
					cpuKib = tmp
				default:
					// Ignore any remaining fields - they are not in most untagged data.
				}
				untaggedPosition++
				matched = true
				anyFields = true

			case taggedFormat:
				if eqloc == CsvEqSentinel {
					// Invalid field syntax: Drop the field but keep the record
					log.Printf(
						"Dropping field with bad form: %s", tokenizer.BufSubstring(start, lim),
					)
					discarded++
					continue FieldLoop
				}

				// The first two characters will always be present because eqloc >= 1.
				switch tokenizer.BufAt(start) {
				case 'c':
					switch tokenizer.BufAt(start + 1) {
					case 'o':
						if val, ok := match(tokenizer, start, lim, eqloc, "cores"); ok {
							var tmp uint64
							tmp, err = strconv.ParseUint(val, 10, 64)
							numCores = uint32(tmp)
							matched = true
						}
					case 'm':
						if val, ok := match(tokenizer, start, lim, eqloc, "cmd"); ok {
							command = ustrs.Alloc(val)
							matched = true
						}
					case 'p':
						if val, ok := match(tokenizer, start, lim, eqloc, "cpu%"); ok {
							var tmp float64
							tmp, err = strconv.ParseFloat(val, 64)
							cpuPct = float32(tmp)
							matched = true
						} else if val, ok := match(tokenizer, start, lim, eqloc, "cpukib"); ok {
							cpuKib, err = strconv.ParseUint(val, 10, 64)
							matched = true
						} else if val, ok := match(tokenizer, start, lim, eqloc, "cputime_sec"); ok {
							cpuTimeSec, err = strconv.ParseUint(val, 10, 64)
							matched = true
						}
					}
				case 'g':
					if val, ok := match(tokenizer, start, lim, eqloc, "gpu%"); ok {
						var tmp float64
						tmp, err = strconv.ParseFloat(val, 64)
						gpuPct = float32(tmp)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "gpumem%"); ok {
						var tmp float64
						tmp, err = strconv.ParseFloat(val, 64)
						gpuMemPct = float32(tmp)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "gpukib"); ok {
						gpuKib, err = strconv.ParseUint(val, 10, 64)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "gpufail"); ok {
						var tmp uint64
						tmp, err = strconv.ParseUint(val, 10, 64)
						gpuFail = uint8(tmp)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "gpus"); ok {
						gpus, err = NewGpuSet(val)
						haveGpus = true
						matched = true
					}
				case 'h':
					if val, ok := match(tokenizer, start, lim, eqloc, "host"); ok {
						hostname = ustrs.Alloc(val)
						matched = true
					}
				case 'j':
					if val, ok := match(tokenizer, start, lim, eqloc, "job"); ok {
						var tmp uint64
						tmp, err = strconv.ParseUint(val, 10, 64)
						jobId = uint32(tmp)
						matched = true
					}
				case 'm':
					if val, ok := match(tokenizer, start, lim, eqloc, "memtotalkib"); ok {
						memTotalKib, err = strconv.ParseUint(val, 10, 64)
						matched = true
					}
				case 'p':
					if val, ok := match(tokenizer, start, lim, eqloc, "pid"); ok {
						var tmp uint64
						tmp, err = strconv.ParseUint(val, 10, 64)
						pid = uint32(tmp)
						matched = true
					}
				case 'r':
					// TODO: Switch on second letter?
					if val, ok := match(tokenizer, start, lim, eqloc, "rssanonkib"); ok {
						rssAnonKib, err = strconv.ParseUint(val, 10, 64)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "rolledup"); ok {
						var tmp uint64
						tmp, err = strconv.ParseUint(val, 10, 64)
						rolledup = uint32(tmp)
						matched = true
					}
				case 't':
					if val, ok := match(tokenizer, start, lim, eqloc, "time"); ok {
						var tmp time.Time
						tmp, err = time.Parse(time.RFC3339Nano, val)
						timestamp = tmp.Unix()
						matched = true
					}
				case 'u':
					if val, ok := match(tokenizer, start, lim, eqloc, "user"); ok {
						user = ustrs.Alloc(val)
						matched = true
					}
				case 'v':
					if val, ok := match(tokenizer, start, lim, eqloc, "v"); ok {
						version = ustrs.Alloc(val)
						matched = true
					}
				}
				if !matched {
					log.Printf(
						"Dropping field with unknown name: %s", tokenizer.BufSubstring(start, eqloc-1),
					)
					if err == nil {
						discarded++
					}
				}
				if err != nil {
					log.Printf(
						"Dropping record with illegal/unparseable value: %s %v",
						tokenizer.BufSubstring(start, lim),
						err,
					)
					discarded++
					tokenizer.ScanEol()
					continue LineLoop
				}
				anyFields = true

			default:
				panic("Unexpected state")
			}
		}

		if !anyFields {
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
				irritants += "timestamp "
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
			log.Printf("Dropping record with missing mandatory field(s): %s", irritants)
			discarded++
			continue LineLoop
		}

		// Fill in default data for optional fields.  Keep this code in sync with initialization
		// above!
		if numCores == math.MaxUint32 {
			numCores = 0
		}
		if memTotalKib == math.MaxUint64 {
			memTotalKib = 0
		}
		if jobId == math.MaxUint32 {
			jobId = 0
		}
		if pid == math.MaxUint32 {
			pid = 0
		}
		if cpuPct == math.MaxFloat32 {
			cpuPct = 0
		}
		if cpuKib == math.MaxUint64 {
			cpuKib = 0
		}
		if rssAnonKib == math.MaxUint64 {
			rssAnonKib = 0
		}
		if !haveGpus {
			gpus = EmptyGpuSet()
		}
		if gpuPct == math.MaxFloat32 {
			gpuPct = 0
		}
		if gpuMemPct == math.MaxFloat32 {
			gpuMemPct = 0
		}
		if gpuKib == math.MaxUint64 {
			gpuKib = 0
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

		if command == heartbeat {
			heartbeats = append(heartbeats, &SonarHeartbeat{
				Version:   version,
				Timestamp: timestamp,
				Host:      hostname,
			})
		} else {
			readings = append(readings, &SonarReading{
				Version:     version,
				Timestamp:   timestamp,
				Host:        hostname,
				Cores:       numCores,
				MemtotalKib: memTotalKib,
				User:        user,
				Pid:         pid,
				Job:         jobId,
				Cmd:         command,
				CpuPct:      cpuPct,
				CpuKib:      cpuKib,
				RssAnonKib:  rssAnonKib,
				Gpus:        gpus,
				GpuPct:      gpuPct,
				GpuMemPct:   gpuMemPct,
				GpuKib:      gpuKib,
				GpuFail:     gpuFail,
				CpuTimeSec:  cpuTimeSec,
				Rolledup:    rolledup,
			})
		}
	}

	err = nil
	return
}

func match(tokenizer *CsvTokenizer, start, lim, eqloc int, tag string) (string, bool) {
	if tokenizer.MatchTag(tag, start, eqloc) {
		return tokenizer.BufSubstring(eqloc, lim), true
	}
	return "", false
}

func (r *SonarReading) Csvnamed() []byte {
	var bw bytes.Buffer
	fields := []string{
		fmt.Sprintf("v=%v", r.Version),
		fmt.Sprintf("time=%s", time.Unix(r.Timestamp, 0).Format(time.RFC3339)),
		fmt.Sprintf("host=%v", r.Host),
		fmt.Sprintf("user=%v", r.User),
		fmt.Sprintf("cmd=%v", r.Cmd),
	}
	if r.Cores > 0 {
		fields = append(fields, fmt.Sprintf("cores=%d", r.Cores))
	}
	if r.MemtotalKib > 0 {
		fields = append(fields, fmt.Sprintf("memtotalkib=%d", r.MemtotalKib))
	}
	if r.Job > 0 {
		fields = append(fields, fmt.Sprintf("job=%d", r.Job))
	}
	if r.Pid > 0 {
		fields = append(fields, fmt.Sprintf("pid=%d", r.Pid))
	}
	if r.CpuPct > 0 {
		fields = append(fields, fmt.Sprintf("cpu%%=%g", r.CpuPct))
	}
	if r.CpuKib > 0 {
		fields = append(fields, fmt.Sprintf("cpukib=%d", r.CpuKib))
	}
	if r.RssAnonKib > 0 {
		fields = append(fields, fmt.Sprintf("rssanonkib=%d", r.RssAnonKib))
	}
	if !r.Gpus.IsEmpty() {
		fields = append(fields, fmt.Sprintf("gpus=%v", r.Gpus))
	}
	if r.GpuPct > 0 {
		fields = append(fields, fmt.Sprintf("gpu%%=%g", r.GpuPct))
	}
	if r.GpuMemPct > 0 {
		fields = append(fields, fmt.Sprintf("gpumem%%=%g", r.GpuMemPct))
	}
	if r.GpuKib > 0 {
		fields = append(fields, fmt.Sprintf("gpukib=%d", r.GpuKib))
	}
	if r.GpuFail != 0 {
		fields = append(fields, fmt.Sprintf("gpufail=%d", r.GpuFail))
	}
	if r.CpuTimeSec > 0 {
		fields = append(fields, fmt.Sprintf("cputime_sec=%d", r.CpuTimeSec))
	}
	if r.Rolledup > 0 {
		fields = append(fields, fmt.Sprintf("rolledup=%d", r.Rolledup))
	}
	csvw := csv.NewWriter(&bw)
	csvw.Write(fields)
	csvw.Flush()
	return bw.Bytes()
}

func (r *SonarHeartbeat) Csvnamed() []byte {
	var bw bytes.Buffer
	csvw := csv.NewWriter(&bw)
	csvw.Write([]string{
		fmt.Sprintf("v=%v", r.Version),
		fmt.Sprintf("time=%s", time.Unix(r.Timestamp, 0).Format(time.RFC3339)),
		fmt.Sprintf("host=%v", r.Host),
		fmt.Sprintf("user=_sonar_"),
		fmt.Sprintf("cmd=_heartbeat_"),
	})
	csvw.Flush()
	return bw.Bytes()
}

// Given one line of text on free csv format, return the pairs of field names and values.
//
// Errors:
// - If the CSV reader returns an error err, returns (nil, err), including io.EOF.
// - If any field is seen not to have a field name, return (fields, ErrNoName) with
//   fields that were valid.

func GetCsvFields(text string) (map[string]string, error) {
	rdr := csv.NewReader(strings.NewReader(text))
	rdr.FieldsPerRecord = -1 // Free form, though should not matter
	fields, err := rdr.Read()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, f := range fields {
		ix := strings.IndexByte(f, '=')
		if ix == -1 {
			err = ErrNoName
			continue
		}
		// TODO: I guess we should detect duplicates
		result[f[0:ix]] = f[ix+1:]
	}
	return result, err
}

var (
	ErrNoName = errors.New("CSV field without a field name")
)
