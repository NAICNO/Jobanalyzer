package sonarlog

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
)

// Sonar CSV data intermingle readings and heartbeats.  Read a stream of them, parse them into
// separate buckets and return the buckets.  Returns the number of benign errors, and non-nil error
// if non-benign error.  The records are in the order they appear in the input.

func ParseSonarCsvnamed(input io.Reader) ([]*SonarReading, []*SonarHeartbeat, int, error) {
	rdr := csv.NewReader(input)
	// CSV rows are arbitrarily wide and possibly uneven.
	rdr.FieldsPerRecord = -1
	badRecords := 0
	readings := make([]*SonarReading, 0)
	heartbeats := make([]*SonarHeartbeat, 0)
outerLoop:
	for {
		fields, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, 0, err
		}
		r := new(SonarReading)
		for _, f := range fields {
			ix := strings.IndexByte(f, '=')
			if ix == -1 {
				// Illegal syntax, just drop the field.
				badRecords++
				continue
			}
			val := f[ix+1:]
			var err error
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
			default:
				// Illegal field, just drop it
				badRecords++
			}
			if err != nil {
				// Illegal value in known field, drop the record
				badRecords++
				continue outerLoop
			}
		}
		if r.Version == "" || r.Timestamp == "" || r.Host == "" || r.Cmd == "" {
			// Missing required fields, drop the record
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
			if r.User == "" {
				// Missing required field, drop the record
				badRecords++
				continue outerLoop
			}
			readings = append(readings, r)
		}
	}
	return readings, heartbeats, badRecords, nil
}

