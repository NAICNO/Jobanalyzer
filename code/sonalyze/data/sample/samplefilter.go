// Sample record filter.  This is performance-sensitive.

package sample

import (
	"errors"
	"fmt"
	"math"
	"os"

	"go-utils/config"
	"go-utils/hostglob"
	umaps "go-utils/maps"
	uslices "go-utils/slices"

	. "sonalyze/common"
	"sonalyze/data/common"
	"sonalyze/db/repr"
)

// The sample.SampleFilter will be applied to individual records and must be true for records to be
// included and false for all others.  It may in principle be applied at any point in the ingestion
// pipeline but at this time it is applied after raw ingestion and is therefore not a part of the
// database layer.
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
// function for filtering.  (If the DB layer were to do any filtering by itself it would be under no
// obligation to use this type of filter function.)
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

func InstantiateSampleFilter(recordFilter *SampleFilter) func(*repr.Sample) bool {
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

	return func(e *repr.Sample) bool {
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
				if !recordFilter.IncludeHosts.Match(e.Hostname.String()) {
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

type QueryFilter struct {
	common.QueryFilter
	AllUsers              bool
	SkipSystemUsers       bool
	ExcludeSystemCommands bool
	ExcludeHeartbeat      bool
	ExcludeSystemJobs     bool
	User                  []string
	ExcludeUser           []string
	Command               []string
	ExcludeCommand        []string
	Job                   []uint32
	ExcludeJob            []uint32
}

func BuildSampleFilter(
	cfg *config.ClusterConfig,
	filter QueryFilter,
	verbose bool,
) (
	*Hosts,
	*SampleFilter,
	error,
) {
	// Included host set, empty means "all"

	includeHosts, err := NewHosts(true, filter.Host)
	if err != nil {
		return nil, nil, err
	}

	// Included job numbers, empty means "all"

	includeJobs := make(map[uint32]bool)
	for _, j := range filter.Job {
		includeJobs[j] = true
	}

	// Excluded job numbers.

	excludeJobs := make(map[uint32]bool)
	for _, j := range filter.ExcludeJob {
		excludeJobs[j] = true
	}

	// Included users, empty means "all"

	allUsers := filter.AllUsers
	includeUsers := make(map[Ustr]bool)
	if len(filter.User) > 0 {
		allUsers = false
		for _, u := range filter.User {
			if u == "-" {
				allUsers = true
				break
			}
		}
		if !allUsers {
			for _, u := range filter.User {
				includeUsers[StringToUstr(u)] = true
			}
		}
	} else if allUsers {
		// Everyone, so do nothing
	} else {
		// LOGNAME is Posix but may be limited to a user being "logged in"; USER is BSD and a bit
		// more general supposedly.  We prefer the former but will settle for the latter, in
		// particular, Github actions has only USER.
		if name := os.Getenv("LOGNAME"); name != "" {
			includeUsers[StringToUstr(name)] = true
		} else if name := os.Getenv("USER"); name != "" {
			includeUsers[StringToUstr(name)] = true
		} else {
			return nil, nil, errors.New("Not able to determine user, none given and $LOGNAME and $USER are empty")
		}
	}

	// Excluded users.

	excludeUsers := make(map[Ustr]bool)
	for _, u := range filter.ExcludeUser {
		excludeUsers[StringToUstr(u)] = true
	}

	if filter.SkipSystemUsers {
		// This list needs to be configurable somehow, but isn't so in the Rust version either.
		excludeUsers[StringToUstr("root")] = true
		excludeUsers[StringToUstr("zabbix")] = true
	}

	// Included commands.

	includeCommands := make(map[Ustr]bool)
	for _, command := range filter.Command {
		includeCommands[StringToUstr(command)] = true
	}

	// Excluded commands.

	excludeCommands := make(map[Ustr]bool)
	for _, command := range filter.ExcludeCommand {
		excludeCommands[StringToUstr(command)] = true
	}

	if filter.ExcludeSystemCommands {
		// This list needs to be configurable somehow, but isn't so in the Rust version either.
		excludeCommands[StringToUstr("bash")] = true
		excludeCommands[StringToUstr("zsh")] = true
		excludeCommands[StringToUstr("sshd")] = true
		excludeCommands[StringToUstr("tmux")] = true
		excludeCommands[StringToUstr("systemd")] = true
	}

	// Skip heartbeat records?  It's probably OK to filter only by command name, since we're
	// currently doing full-command-name matching.

	if filter.ExcludeHeartbeat {
		excludeCommands[StringToUstr("_heartbeat_")] = true
	}

	// System configuration additions, if available

	if cfg != nil {
		for _, user := range cfg.ExcludeUser {
			excludeUsers[StringToUstr(user)] = true
		}
	}

	// Record filtering logic is the same for all commands.  The record filter can use only raw
	// ingested data, it can be applied at any point in the pipeline.  It *must* be thread-safe.

	excludeSystemJobs := filter.ExcludeSystemJobs
	haveFrom := filter.HaveFrom
	haveTo := filter.HaveTo
	var from int64 = 0
	if haveFrom {
		from = filter.FromDate.Unix()
	}
	var to int64 = math.MaxInt64
	if haveTo {
		to = filter.ToDate.Unix()
	}
	var minPid uint32
	if excludeSystemJobs {
		minPid = 1000
	}

	var recordFilter = &SampleFilter{
		IncludeUsers:    includeUsers,
		IncludeHosts:    includeHosts.HostnameGlobber(),
		IncludeJobs:     includeJobs,
		IncludeCommands: includeCommands,
		ExcludeUsers:    excludeUsers,
		ExcludeJobs:     excludeJobs,
		ExcludeCommands: excludeCommands,
		MinPid:          minPid,
		From:            from,
		To:              to,
	}

	if verbose {
		if haveFrom {
			Log.Infof("Including records starting on or after %s", filter.FromDate)
		}
		if haveTo {
			Log.Infof("Including records ending on or before %s", filter.ToDate)
		}
		if len(includeUsers) > 0 {
			Log.Infof("Including records with users %s", umaps.Values(includeUsers))
		}
		if !includeHosts.IsEmpty() {
			Log.Infof("Including records with hosts matching %s", includeHosts)
		}
		if len(includeJobs) > 0 {
			Log.Infof(
				"Including records with job id matching %s",
				uslices.Map(umaps.Keys(includeJobs), func(x uint32) string {
					return fmt.Sprint(x)
				}))
		}
		if len(includeCommands) > 0 {
			Log.Infof("Including records with commands matching %s", umaps.Keys(includeCommands))
		}
		if len(excludeUsers) > 0 {
			Log.Infof("Excluding records with users matching %s", umaps.Keys(excludeUsers))
		}
		if len(excludeJobs) > 0 {
			Log.Infof(
				"Excluding records with job ids matching %s",
				uslices.Map(umaps.Keys(excludeJobs), func(x uint32) string {
					return fmt.Sprint(x)
				}))
		}
		if len(excludeCommands) > 0 {
			Log.Infof("Excluding records with commands matching %s", umaps.Keys(excludeCommands))
		}
		if excludeSystemJobs {
			Log.Infof("Excluding records with PID < 1000")
		}
	}

	return includeHosts, recordFilter, nil
}
