// Fast, non-allocating sonar-log CSV parser.

package sonarlog

import (
	"bytes"
	"errors"
	"go-utils/gpuset"
	"io"
	"log"
	"math"
	"time"

	. "sonalyze/common"
)

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
	readings []*Sample,
	discarded int,
	err error,
) {
	const (
		unknownFormat = iota
		untaggedFormat
		taggedFormat
	)

	readings = make([]*Sample, 0)
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
			memTotalKib      uint64  = math.MaxUint64
			user                     = UstrEmpty
			pid              uint32  = math.MaxUint32
			jobId            uint32  = math.MaxUint32
			command                  = UstrEmpty
			cpuPct           float32 = math.MaxFloat32
			cpuKib           uint64  = math.MaxUint64
			rssAnonKib       uint64  = math.MaxUint64
			gpus                     = gpuset.EmptyGpuSet()
			haveGpus                 = false
			gpuPct           float32 = math.MaxFloat32
			gpuMemPct        float32 = math.MaxFloat32
			gpuKib           uint64  = math.MaxUint64
			gpuFail          uint8   = math.MaxUint8
			cpuTimeSec       uint64  = math.MaxUint64
			rolledup         uint32  = math.MaxUint32
			format                   = unknownFormat
			untaggedPosition         = 0
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
						discarded++
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
						discarded++
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
						discarded++
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
						discarded++
						tokenizer.ScanEol()
						continue LineLoop
					}
					cpuPct = float32(tmp)
				case 7:
					var tmp uint64
					tmp, err = parseUint(val)
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

			case taggedFormat:
				// NOTE, in error cases below we don't extract the offending field b/c it seems the
				// optimizer will hoist the (technically effect-free) extraction out of the parsing
				// switch and slow everything down tremendously.

				if eqloc == CsvEqSentinel {
					// Invalid field syntax: Drop the field but keep the record
					if verbose {
						log.Printf(
							"Dropping field with bad form: %s",
							"(elided)", /*tokenizer.BufSubstringSlow(start, lim), - see NOTE above*/
						)
					}
					discarded++
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
									cpuKib, err = parseUint(val)
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
						case 'k':
							if val, ok := match(tokenizer, start, lim, eqloc, "gpukib"); ok {
								gpuKib, err = parseUint(val)
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
				case 'm':
					if val, ok := match(tokenizer, start, lim, eqloc, "memtotalkib"); ok {
						memTotalKib, err = parseUint(val)
						matched = true
					}
				case 'p':
					if val, ok := match(tokenizer, start, lim, eqloc, "pid"); ok {
						var tmp uint64
						tmp, err = parseUint(val)
						pid = uint32(tmp)
						matched = true
					}
				case 'r':
					if val, ok := match(tokenizer, start, lim, eqloc, "rssanonkib"); ok {
						rssAnonKib, err = parseUint(val)
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
						log.Printf(
							"Dropping field with unknown name: %s",
							"(elided)", /* tokenizer.BufSubstringSlow(start, eqloc-1), -
							   see NOTE above */
						)
					}
					if err == nil {
						discarded++
					}
				}
				if err != nil {
					if verbose {
						log.Printf(
							"Dropping record with illegal/unparseable value: %s %v",
							"(elided)", /*tokenizer.BufSubstringSlow(start, lim), - see NOTE above */
							err,
						)
					}
					discarded++
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
				log.Printf(
					"Dropping untagged record with missing fields, got only %d fields",
					untaggedPosition,
				)
			}
			discarded++
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
			if verbose {
				log.Printf("Dropping record with missing mandatory field(s): %s", irritants)
			}
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
			gpus = gpuset.EmptyGpuSet()
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

		flags := uint8(0)
		if command == heartbeat {
			flags |= FlagHeartbeat
		}
		readings = append(readings, &Sample{
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
			Flags:       flags,
		})
	}

	err = nil
	return
}

func match(tokenizer *CsvTokenizer, start, lim, eqloc int, tag string) ([]byte, bool) {
	if tokenizer.MatchTag(tag, start, eqloc) {
		return tokenizer.BufSlice(eqloc, lim), true
	}
	return nil, false
}

func parseUint(bs []byte) (uint64, error) {
	var n uint64
	if len(bs) == 0 {
		return 0, errors.New("Empty")
	}
	for _, c := range bs {
		if c < '0' || c > '9' {
			return 0, errors.New("Not a digit")
		}
		m := n*10 + uint64(c-'0')
		if m < n {
			return 0, errors.New("Out of range")
		}
		n = m
	}
	return n, nil
}

// A faster number parser operating on byte slices.
//
// This is primitive and - except for NaN and Infinity - handles only simple unsigned numbers with a
// fraction, no exponent.  Accuracy is not great either.  But it's good enough for the Sonar output,
// which should have no exponentials, require low accuracy, and only occasionally - in older,
// buggier data - has NaN and Infinity.
//
// According to documentation, strconv.ParseFloat() accepts nan, inf, +inf, -inf, infinity,
// +infinity and -infinity, case-insensitively.
//
// Based on experimentation, the rust to_string() formatter will produce "NaN", "inf" and "-inf",
// with that capitalization.
//
// Based on experimentation, the Go formatter produces "NaN", "+Inf" and "-Inf".
func parseFloat(bs []byte, filterInfNaN bool) (float64, error) {
	// Canonical code
	// x, err := strconv.ParseFloat(string(bs), 64)
	// if err != nil {
	// 	return 0, err
	// }
	// if filterInfNaN && (math.IsInf(x, 0) || math.IsNaN(x)) {
	// 	return 0, errors.New("Infinity / NaN")
	// }
	// return x, nil
	var n float64
	if len(bs) == 0 {
		return 0, errors.New("Empty")
	}
	switch bs[0] {
	case '-':
		// No negative numbers
		return 0, errors.New("Not a digit")
	case '+':
		if bytes.EqualFold(bs, []byte{'+', 'i', 'n', 'f', 'i', 'n', 'i', 't', 'y'}) ||
			bytes.EqualFold(bs, []byte{'+', 'i', 'n', 'f'}) {
			if filterInfNaN {
				return 0, errors.New("Infinity")
			}
			return math.Inf(1), nil
		}
		return 0, errors.New("Not a digit")
	case 'i', 'I':
		if bytes.EqualFold(bs, []byte{'i', 'n', 'f', 'i', 'n', 'i', 't', 'y'}) ||
			bytes.EqualFold(bs, []byte{'i', 'n', 'f'}) {
			if filterInfNaN {
				return 0, errors.New("Infinity")
			}
			return math.Inf(1), nil
		}
		return 0, errors.New("Not a digit")
	case 'n', 'N':
		if bytes.EqualFold(bs, []byte{'n', 'a', 'n'}) {
			if filterInfNaN {
				return 0, errors.New("NaN")
			}
			return math.NaN(), nil
		}
		return 0, errors.New("Not a digit")
	}
	i := 0
	for ; i < len(bs); i++ {
		c := bs[i]
		if c == '.' {
			break
		}
		if c < '0' || c > '9' {
			return 0, errors.New("Not a digit")
		}
		n = n*10 + float64(c-'0')
	}
	if i < len(bs) {
		if bs[i] != '.' {
			return 0, errors.New("Only decimal point allowed")
		}
		i++
		if i == len(bs) {
			return 0, errors.New("Empty fraction")
		}
		f := 0.1
		for ; i < len(bs); i++ {
			c := bs[i]
			if c < '0' || c > '9' {
				return 0, errors.New("Not a digit")
			}
			n += float64(c-'0') * f
			f *= 0.1
		}
	}
	return n, nil
}
