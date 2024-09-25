// Sample record filter.  This is performance-sensitive.

package db

import (
	"go-utils/hostglob"

	. "sonalyze/common"
)

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
// obligation to use this type of filter function.
//
// The recordFilter or its parts *MUST NOT* be modified by the caller subsequently, as parts of it
// may be retained by the filter function and may be used on concurrent threads.  For all practical
// purposes, the filter function takes (shared) ownership of the recordFilter.  The filter function
// will not update the recordFilter or its parts.
//
// The filter function will be thread-safe and mostly non-locking and non-contending.
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
//   the regex engine, impact TBD but expected to be minor.)
//
// The SampleFilter has a number of components, many of which are empty in most cases.  Filtering
// can be a major expense, so performance is important.  There are numerous options for implementing
// this:
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
// It is beneficial to specialize for sets-of-size-one to avoid hashing and lookup overhead, this
// specialization is possible for the later implementations in the list.

// This is the code for "simple conjunction of filters", preserved here for posterity.
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

// Simple bytecode compiler and interpreter.  For typical filters this is much faster than the above
// code, some sample queries against two months of Fox data run in about half the time.

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
	// "_heartbeat_" command name.  This type of merging would only require one operand, the Job ID.
	// Alternatively we could specialize pairs with two operands if both operands fit in the 59 bits
	// available (basically always).  Implementing this optimization amounts to running a peephole
	// optimizer on the generated instructions.  It's not obvious that it would save much time.

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
