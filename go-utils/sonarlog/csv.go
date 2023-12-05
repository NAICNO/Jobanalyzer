package sonarlog

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"
)

// Sonar "csvnamed" data intermingle readings and heartbeats.  Read a stream of records, parse them
// into separate buckets and return the buckets.  Returns the number of benign errors, and non-nil
// error if non-benign error.  The records in the buckets are in the order they appear in the input.
//
// Note wrt parsing floats: According to documentation, strconv.ParseFloat() accepts nan, inf, +inf,
// -inf, infinity, +infinity and -infinity, case-insensitively.  Based on experimentation, the rust
// to_string() formatter will produce "NaN", "inf" and "-inf", with that capitalization (weird).  So
// ingesting CSV data from Rust should not be a problem.

func ParseSonarCsvnamed(input io.Reader) (readings []*SonarReading, heartbeats []*SonarHeartbeat, badRecords int, err error) {
	rdr := csv.NewReader(input)
	// CSV rows are arbitrarily wide and possibly uneven.
	rdr.FieldsPerRecord = -1
	readings = make([]*SonarReading, 0)
	heartbeats = make([]*SonarHeartbeat, 0)
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
		r := new(SonarReading)
		for _, f := range fields {
			ix := strings.IndexByte(f, '=')
			if ix == -1 {
				log.Printf("Dropping field with illegal syntax: %s", f)
				badRecords++
				continue
			}
			val := f[ix+1:]
			switch f[:ix] {
			case "v":
				r.Version = val
			case "time":
				r.Timestamp = val
			case "host":
				r.Host = val
			case "cores":
				r.Cores, err = strconv.ParseUint(val, 10, 64)
			case "user":
				r.User = val
			case "job":
				r.Job, err = strconv.ParseUint(val, 10, 64)
			case "pid":
				r.Pid, err = strconv.ParseUint(val, 10, 64)
			case "cmd":
				r.Cmd = val
			case "cpu%":
				r.CpuPct, err = strconv.ParseFloat(val, 64)
			case "cpukib":
				r.CpuKib, err = strconv.ParseUint(val, 10, 64)
			case "gpus":
				// We don't validate the gpu syntax here
				r.Gpus = val
			case "gpu%":
				r.GpuPct, err = strconv.ParseFloat(val, 64)
			case "gpumem%":
				r.GpuMemPct, err = strconv.ParseFloat(val, 64)
			case "gpukib":
				r.GpuKib, err = strconv.ParseUint(val, 10, 64)
			case "cputime_sec":
				r.CpuTimeSec, err = strconv.ParseUint(val, 10, 64)
			case "rolledup":
				r.Rolledup, err = strconv.ParseUint(val, 10, 64)
			default:
				log.Printf("Dropping field with unknown name: %s", f)
				badRecords++
			}
			if err != nil {
				log.Printf("Dropping record with illegal/unparseable value: %s", f)
				badRecords++
				continue outerLoop
			}
		}

		irritants := ""
		if r.Version == "" || r.Timestamp == "" || r.Host == "" || r.Cmd == "" {
			if r.Version == "" {
				irritants += "version "
			}
			if r.Timestamp == "" {
				irritants += "timestamp "
			}
			if r.Host == "" {
				irritants += "host "
			}
			if r.Cmd == "" {
				irritants += "cmd "
			}
		}
		if r.Cmd != "_heartbeat_" && r.User == "" {
			irritants += "user "
		}
		if irritants != "" {
			log.Printf("Dropping record with missing mandatory field(s): %s", irritants)
			badRecords++
			continue outerLoop
		}

		if r.Cmd == "_heartbeat_" {
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

