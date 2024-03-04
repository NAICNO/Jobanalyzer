package sonarlog

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

func match(f, k string) (string, bool) {
	if len(f) <= len(k) || f[len(k)] != '=' {
		return "", false
	}
	for i, c := range k {
		if rune(f[i]) != c {
			return "", false
		}
	}
	return f[len(k)+1:], true
}

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
	badRecords int,
	err error,
) {

	rdr := csv.NewReader(input)
	// CSV rows are arbitrarily wide and possibly uneven.
	rdr.FieldsPerRecord = -1
	readings = make([]*SonarReading, 0)
	heartbeats = make([]*SonarHeartbeat, 0)
	v060 := ustrs.Alloc("0.6.0")
	heartbeat := ustrs.Alloc("_heartbeat_")
outerLoop:
	for {
		var fields []string
		fields, err = rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		if len(fields) == 0 {
			badRecords++
			continue outerLoop
		}
		r := new(SonarReading)
		// If the first field starts with a '2' then this is the old untagged format because that's
		// the first byte in an untagged timestamp, and no tags start with that value.
		if []byte(fields[0])[0] == '2' {
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

			if len(fields) < 8 {
				badRecords++
				continue outerLoop
			}

			r.Version = v060
			ts, err := time.Parse(time.RFC3339Nano, fields[0])
			if err != nil {
				badRecords++
				continue outerLoop
			}
			r.Timestamp = ts.Unix()
			r.Host = ustrs.Alloc(fields[1])
			cores, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				badRecords++
				continue outerLoop
			}
			r.Cores = uint32(cores)
			r.User = ustrs.Alloc(fields[3])
			jobno, err := strconv.ParseUint(fields[4], 10, 64)
			if err != nil {
				badRecords++
				continue outerLoop
			}
			r.Job = uint32(jobno)
			r.Pid = r.Job
			r.Cmd = ustrs.Alloc(fields[5])
			cpupct, err := strconv.ParseFloat(fields[6], 64)
			if err != nil {
				badRecords++
				continue outerLoop
			}
			r.CpuPct = float32(cpupct)
			r.CpuKib, err = strconv.ParseUint(fields[7], 10, 64)
			if err != nil {
				badRecords++
				continue outerLoop
			}
			// Skip any remaining fields - they are not in most untagged data.
		} else {
			for _, f := range fields {
				var cores, jobno, pidno, gpufail, rolledup uint64
				var cpupct, gpupct, gpumempct float64
				var ts time.Time
				matched := false
				if len(f) < 2 {
					badRecords++
					continue outerLoop
				}
				switch f[0] {
				case 'c':
					switch f[1] {
					case 'o':
						if val, ok := match(f, "cores"); ok {
							cores, err = strconv.ParseUint(val, 10, 64)
							r.Cores = uint32(cores)
							matched = true
						}
					case 'm':
						if val, ok := match(f, "cmd"); ok {
							r.Cmd = ustrs.Alloc(val)
							matched = true
						}
					case 'p':
						if val, ok := match(f, "cpu%"); ok {
							cpupct, err = strconv.ParseFloat(val, 64)
							r.CpuPct = float32(cpupct)
							matched = true
						} else if val, ok := match(f, "cpukib"); ok {
							r.CpuKib, err = strconv.ParseUint(val, 10, 64)
							matched = true
						} else if val, ok := match(f, "cputime_sec"); ok {
							r.CpuTimeSec, err = strconv.ParseUint(val, 10, 64)
							matched = true
						}
					}
				case 'g':
					if val, ok := match(f, "gpu%"); ok {
						gpupct, err = strconv.ParseFloat(val, 64)
						r.GpuPct = float32(gpupct)
						matched = true
					} else if val, ok := match(f, "gpumem%"); ok {
						gpumempct, err = strconv.ParseFloat(val, 64)
						r.GpuMemPct = float32(gpumempct)
						matched = true
					} else if val, ok := match(f, "gpukib"); ok {
						r.GpuKib, err = strconv.ParseUint(val, 10, 64)
						matched = true
					} else if val, ok := match(f, "gpufail"); ok {
						gpufail, err = strconv.ParseUint(val, 10, 64)
						r.GpuFail = uint8(gpufail)
						matched = true
					} else if val, ok := match(f, "gpus"); ok {
						r.Gpus, err = NewGpuSet(val)
						matched = true
					}
				case 'h':
					if val, ok := match(f, "host"); ok {
						r.Host = ustrs.Alloc(val)
						matched = true
					}
				case 'j':
					if val, ok := match(f, "job"); ok {
						jobno, err = strconv.ParseUint(val, 10, 64)
						r.Job = uint32(jobno)
						matched = true
					}
				case 'm':
					if val, ok := match(f, "memtotalkib"); ok {
						r.MemtotalKib, err = strconv.ParseUint(val, 10, 64)
						matched = true
					}
				case 'p':
					if val, ok := match(f, "pid"); ok {
						pidno, err = strconv.ParseUint(val, 10, 64)
						r.Pid = uint32(pidno)
						matched = true
					}
				case 'r':
					if val, ok := match(f, "rssanonkib"); ok {
						r.RssAnonKib, err = strconv.ParseUint(val, 10, 64)
						matched = true
					} else if val, ok := match(f, "rolledup"); ok {
						rolledup, err = strconv.ParseUint(val, 10, 64)
						r.Rolledup = uint32(rolledup)
						matched = true
					}
				case 't':
					if val, ok := match(f, "time"); ok {
						// This is really the format we use in the logs, but the nano part is often
						// omitted by our formatters:
						//
						//  "2006-01-02T15:04:05.999999999-07:00"
						//
						// RFC3339Nano handles +/- for the tz offset and also will allow the nano
						// part to be missing.
						ts, err = time.Parse(time.RFC3339Nano, val)
						r.Timestamp = ts.Unix()
						matched = true
					}
				case 'u':
					if val, ok := match(f, "user"); ok {
						r.User = ustrs.Alloc(val)
						matched = true
					}
				case 'v':
					if val, ok := match(f, "v"); ok {
						r.Version = ustrs.Alloc(val)
						matched = true
					}
				}
				if !matched {
					log.Printf("Dropping field with unknown name: %s", f)
					badRecords++
				}
				if err != nil {
					log.Printf("Dropping record with illegal/unparseable value: %s %v", f, err)
					badRecords++
					continue outerLoop
				}
			}
		}

		irritants := ""
		if r.Version == UstrEmpty || r.Timestamp == 0 || r.Host == UstrEmpty || r.Cmd == UstrEmpty {
			if r.Version == UstrEmpty {
				irritants += "version "
			}
			if r.Timestamp == 0 {
				irritants += "timestamp "
			}
			if r.Host == UstrEmpty {
				irritants += "host "
			}
			if r.Cmd == UstrEmpty {
				irritants += "cmd "
			}
		}
		if r.Cmd != heartbeat && r.User == UstrEmpty {
			irritants += "user "
		}
		if irritants != "" {
			log.Printf("Dropping record with missing mandatory field(s): %s", irritants)
			badRecords++
			continue outerLoop
		}

		if r.Cmd == heartbeat {
			heartbeats = append(heartbeats, &SonarHeartbeat{
				Version:   r.Version,
				Timestamp: r.Timestamp,
				Host:      r.Host,
			})
		} else {
			readings = append(readings, r)
		}
	}

	err = nil
	return
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
