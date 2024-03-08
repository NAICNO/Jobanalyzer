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

// Sonar "csvnamed" data intermingle readings and heartbeats.  Read a stream of records, parse them
// into separate buckets and return the buckets.  Returns the number of benign errors, and non-nil
// error if non-benign error.  The records in the buckets are in the order they appear in the input.
//
// Note wrt parsing floats: According to documentation, strconv.ParseFloat() accepts nan, inf, +inf,
// -inf, infinity, +infinity and -infinity, case-insensitively.  Based on experimentation, the rust
// to_string() formatter will produce "NaN", "inf" and "-inf", with that capitalization (weird).  So
// ingesting CSV data from Rust should not be a problem.

func ParseSonarCsvnamed(
	input io.Reader,
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
	heartbeat := StringToUstr("_heartbeat_")
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
			var tmp uint64
			var ftmp float64
			var ts time.Time
			val := f[ix+1:]
			switch f[:ix] {
			case "v":
				r.Version = StringToUstr(val)
			case "time":
				// This is really the format we use in the logs, but the nano part is often omitted
				// by our formatters:
				//
				//  "2006-01-02T15:04:05.999999999-07:00"
				//
				// RFC3339Nano handles +/- for the tz offset and also will allow the nano part to be
				// missing.
				ts, err = time.Parse(time.RFC3339Nano, val)
				if err == nil {
					r.Timestamp = ts.Unix()
				}
			case "host":
				r.Host = StringToUstr(val)
			case "cores":
				tmp, err = strconv.ParseUint(val, 10, 64)
				r.Cores = uint32(tmp)
			case "memtotalkib":
				r.MemtotalKib, err = strconv.ParseUint(val, 10, 64)
			case "user":
				r.User = StringToUstr(val)
			case "job":
				tmp, err = strconv.ParseUint(val, 10, 64)
				r.Job = uint32(tmp)
			case "pid":
				tmp, err = strconv.ParseUint(val, 10, 64)
				r.Pid = uint32(tmp)
			case "cmd":
				r.Cmd = StringToUstr(val)
			case "cpu%":
				ftmp, err = strconv.ParseFloat(val, 64)
				r.CpuPct = float32(ftmp)
			case "cpukib":
				r.CpuKib, err = strconv.ParseUint(val, 10, 64)
			case "rssanonkib":
				r.RssAnonKib, err = strconv.ParseUint(val, 10, 64)
			case "gpus":
				// We don't validate the gpu syntax here
				r.Gpus = StringToUstr(val)
			case "gpu%":
				ftmp, err = strconv.ParseFloat(val, 64)
				r.GpuPct = float32(ftmp)
			case "gpumem%":
				ftmp, err = strconv.ParseFloat(val, 64)
				r.GpuMemPct = float32(ftmp)
			case "gpukib":
				r.GpuKib, err = strconv.ParseUint(val, 10, 64)
			case "gpufail":
				tmp, err = strconv.ParseUint(val, 10, 64)
				r.GpuFail = uint8(tmp)
			case "cputime_sec":
				r.CpuTimeSec, err = strconv.ParseUint(val, 10, 64)
			case "rolledup":
				tmp, err = strconv.ParseUint(val, 10, 64)
				r.Rolledup = uint32(tmp)
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
	if r.Gpus != UstrEmpty {
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
