// Fast, non-allocating sonar-log CSV parser.

package db

import (
	"errors"
	"go-utils/gpuset"
	"io"
	"math"
	"slices"
	"time"

	"go-utils/hostglob"

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
	samples []*Sample,
	loadData []*LoadDatum,
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
			ppid             uint32  = math.MaxUint32
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
			load             []byte
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
				case 'l':
					if val, ok := match(tokenizer, start, lim, eqloc, "load"); ok {
						load = slices.Clone(val)
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
					} else if val, ok := match(tokenizer, start, lim, eqloc, "ppid"); ok {
						var tmp uint64
						tmp, err = parseUint(val)
						ppid = uint32(tmp)
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
		if memTotalKib == math.MaxUint64 {
			memTotalKib = 0
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
		samples = append(samples, &Sample{
			Version:     version,
			Timestamp:   timestamp,
			Host:        hostname,
			Cores:       numCores,
			MemtotalKib: memTotalKib,
			User:        user,
			Pid:         pid,
			Ppid:        ppid,
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
		if load != nil {
			loadData = append(loadData, &LoadDatum{
				Timestamp: timestamp,
				Host:      hostname,
				Encoded:   load,
			})
		}
	}

	err = nil
	return
}

// The db.SampleFilter will be applied to individual records and must be true for records to be
// included and false for all others.  It may be applied at any point in the ingestion pipeline.
//
// The fields may all have zero values; however, if `To` is not a precise time then it should be set
// to math.MaxInt64, as zero will mean zero here too.
//
// Maps should only carry `true` values for present elements - these are sets.

type SampleFilter struct {
	IncludeUsers    map[Ustr]bool
	IncludeHosts    *hostglob.HostGlobber
	IncludeJobs     map[uint32]bool
	IncludeCommands map[Ustr]bool
	ExcludeUsers    map[Ustr]bool
	ExcludeJobs     map[uint32]bool
	ExcludeCommands map[Ustr]bool
	MinPid          uint32
	From            int64
	To              int64
}

// Construct an efficient filter function from the sample filter.
//
// This is a convenience function for external code that has a SampleFilter but needs to have a
// function for filtering.  If the DB layer were to do any filtering by itself it would be under no
// obligation to use this.
//
// The recordFilter or its parts *MUST NOT* be modified by the caller subsequently, as parts of it
// may be retained by the filter function and may be used on concurrent threads.  For all practical
// purposes, the returned function takes (shared) ownership of the recordFilter.  The returned
// function however does not update the recordFilter or its part.
//
// The returned function will be thread-safe and mostly non-locking and non-contending.
//
// Notes:
//
// - Go maps are safe for concurrent read access without locking and can be used by the SampleFilter
//   with that restriction.  From https://go.dev/doc/faq#atomic_maps:
//
//     Map access is unsafe only when updates are occurring. As long as all goroutines are only
//     reading—looking up elements in the map, including iterating through it using a for range
//     loop—and not changing the map by assigning to elements or doing deletions, it is safe for
//     them to access the map concurrently without synchronization.
//
//   Almost certainly, the program must be careful to establish a happens-before relationship
//   between map initialization and all map reads for this to be true.  Since the map will likely be
//   broadcast to a bunch of goroutines this will often happen as a matter of course.
//
// - The Sonalyze HostGlobber is thread-safe (but not always contention-free due to shared state in
//   the regex engine, impact TBD.)
//
// The filter has a number of components, many of which are empty in most cases.  Filtering can be a
// major expense, so performance is important.  There are numerous options:
//
// - Simple conjunction of filters in their most straightforward state.
//
// - Ditto, but with fast-checkable flags for each case to avoid doing work, probably nil checks in
//   the implementation will serve the same function as these flags in most cases though, so this
//   would likely be very modest gain at best, over the previous case.
//
// - Special implementations for common or important cases, and a default (one of the above) for the
//   rest, eg, filters that filter "only" on one job ID.  This is surprisingly hard due to how the
//   filters are used in practice to filter eg heartbeat records and from/to ranges: no filters
//   actually filter just on one thing.
//
// - A list of closures or a nested closure, one closure for each attribute to filter by, avoiding
//   code altogether for absent cases.  This adds call overhead per case but may still be a win
//   because many cases will not be tested for.
//
// - A bytecode representation of active filters so that there is no closure call overhead and no
//   redundant tests, but we add dispatch overhead per case again.
//
// It is also beneficial to specialize for sets-of-size-one to avoid hashing and lookup overhead,
// this specialization is possible for the later implementations in the list.

// This is the canonical code, preserved here for posterity.
/*
func InstantiateSampleFilter0(recordFilter *SampleFilter) func(*Sample) bool {
	return func(e *Sample) bool {
		return (len(recordFilter.IncludeUsers) == 0 || recordFilter.IncludeUsers[e.User]) &&
			(recordFilter.IncludeHosts == nil ||
				recordFilter.IncludeHosts.IsEmpty() ||
				recordFilter.IncludeHosts.Match(e.Host.String())) &&
			(len(recordFilter.IncludeJobs) == 0 || recordFilter.IncludeJobs[e.Job]) &&
			(len(recordFilter.IncludeCommands) == 0 || recordFilter.IncludeCommands[e.Cmd]) &&
			!recordFilter.ExcludeUsers[e.User] &&
			!recordFilter.ExcludeJobs[e.Job] &&
			!recordFilter.ExcludeCommands[e.Cmd] &&
			e.Pid >= recordFilter.MinPid &&
			recordFilter.From <= e.Timestamp &&
			e.Timestamp <= recordFilter.To
	}
}
*/

// Simle bytecode interpreter.  For typical filters this is much faster than the above code, some
// sample queries against two months of Fox data run in about half the time.

func InstantiateSampleFilter(recordFilter *SampleFilter) func(*Sample) bool {
	// The filter is a simple bytecode interpreter so as to avoid redundancies and easily specialize
	// fast cases.  Tests are ordered from most to least likely and most to least discriminating.
	//
	// All instructions are 64-bit ints, sometimes there is a payload.  This structure allows us to
	// use a ranged `for` loop to avoid bounds checking, program counters, and end-of-program
	// testing.
	//
	// Tests that are almost always done - from/to filtering here - are outside the loop to avoid
	// the pointless dispatch overhead.
	//
	// TODO: A further optimization here is to merge common pairs of operations to avoid dispatch
	// overhead - for example, a typical filter is to include by Job ID and exclude by the
	// "_heartbeat_" command name.  This type of merging would only require one operand.
	// Alternatively we could specialize pairs with two operands if both operands fit in the 59 bits
	// available (basically always).  Implementing this optimization amounts to running a peephole
	// optimizer on the generated instructions.

	const (
		testIncludeSingleJob uint64 = iota
		testIncludeJobs
		testIncludeSingleUser
		testIncludeUsers
		testIncludeHosts
		testIncludeSingleCommand
		testIncludeCommands
		testExcludeSingleUser
		testExcludeUsers
		testExcludeSingleJob
		testExcludeJobs
		testExcludeSingleCommand
		testExcludeCommands
		testExcludeLowPids
	)

	const (
		// Opcode in low 5 bits
		opMask = 31

		// Operand in high 32 bits, leaving 27 free bits in the middle
		opShift = 32
	)

	instr := make([]uint64, 0)

	switch len(recordFilter.IncludeJobs) {
	case 0:
	case 1:
		var theJob uint32
		for j := range recordFilter.IncludeJobs {
			theJob = j
		}
		instr = append(instr, testIncludeSingleJob|uint64(theJob)<<opShift)
	default:
		instr = append(instr, testIncludeJobs)
	}

	switch len(recordFilter.IncludeUsers) {
	case 0:
	case 1:
		var theUser Ustr
		for u := range recordFilter.IncludeUsers {
			theUser = u
		}
		instr = append(instr, testIncludeSingleUser|uint64(theUser)<<opShift)
	default:
		instr = append(instr, testIncludeUsers)
	}

	if recordFilter.IncludeHosts != nil && !recordFilter.IncludeHosts.IsEmpty() {
		instr = append(instr, testIncludeHosts)
	}

	switch len(recordFilter.IncludeCommands) {
	case 0:
	case 1:
		var theCommand Ustr
		for c := range recordFilter.IncludeCommands {
			theCommand = c
		}
		instr = append(instr, testIncludeSingleCommand|uint64(theCommand)<<opShift)
	default:
		instr = append(instr, testIncludeCommands)
	}

	switch len(recordFilter.ExcludeUsers) {
	case 0:
	case 1:
		var theUser Ustr
		for u := range recordFilter.ExcludeUsers {
			theUser = u
		}
		instr = append(instr, testExcludeSingleUser|uint64(theUser)<<opShift)
	default:
		instr = append(instr, testExcludeUsers)
	}

	switch len(recordFilter.ExcludeJobs) {
	case 0:
	case 1:
		var theJob uint32
		for j := range recordFilter.ExcludeJobs {
			theJob = j
		}
		instr = append(instr, testExcludeSingleJob|uint64(theJob)<<opShift)
	default:
		instr = append(instr, testExcludeJobs)
	}

	switch len(recordFilter.ExcludeCommands) {
	case 0:
	case 1:
		var theCommand Ustr
		for c := range recordFilter.ExcludeCommands {
			theCommand = c
		}
		instr = append(instr, testExcludeSingleCommand|uint64(theCommand)<<opShift)
	default:
		instr = append(instr, testExcludeCommands)
	}

	if recordFilter.MinPid > 0 {
		instr = append(instr, testExcludeLowPids|uint64(recordFilter.MinPid)<<opShift)
	}

	return func(e *Sample) bool {
		for _, op := range instr {
			switch op & opMask {
			case testIncludeSingleJob:
				job := uint32(op >> opShift)
				if e.Job != job {
					return false
				}
			case testIncludeJobs:
				if !recordFilter.IncludeJobs[e.Job] {
					return false
				}
			case testIncludeSingleUser:
				uid := Ustr(uint32(op >> opShift))
				if e.User != uid {
					return false
				}
			case testIncludeUsers:
				if !recordFilter.IncludeUsers[e.User] {
					return false
				}
			case testIncludeHosts:
				if !recordFilter.IncludeHosts.Match(e.Host.String()) {
					return false
				}
			case testIncludeSingleCommand:
				cmd := Ustr(uint32(op >> opShift))
				if e.Cmd != cmd {
					return false
				}
			case testIncludeCommands:
				if !recordFilter.IncludeCommands[e.Cmd] {
					return false
				}
			case testExcludeSingleUser:
				uid := Ustr(uint32(op >> opShift))
				if e.User == uid {
					return false
				}
			case testExcludeUsers:
				if recordFilter.ExcludeUsers[e.User] {
					return false
				}
			case testExcludeSingleJob:
				job := uint32(op >> opShift)
				if e.Job == job {
					return false
				}
			case testExcludeJobs:
				if recordFilter.ExcludeJobs[e.Job] {
					return false
				}
			case testExcludeSingleCommand:
				cmd := Ustr(uint32(op >> opShift))
				if e.Cmd == cmd {
					return false
				}
			case testExcludeCommands:
				if recordFilter.ExcludeCommands[e.Cmd] {
					return false
				}
			case testExcludeLowPids:
				pid := uint32(op >> opShift)
				if e.Pid < pid {
					return false
				}
			}
		}

		// For any selection of note, these will always have to be run, but they will almost always
		// pass due to the structure of the database.  So apply them only at the end.
		return recordFilter.From <= e.Timestamp && e.Timestamp <= recordFilter.To
	}
}
