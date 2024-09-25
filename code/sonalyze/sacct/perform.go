package sacct

import (
	"fmt"
	"io"

	"go-utils/hostglob"
	uslices "go-utils/slices"

	. "sonalyze/common"
	"sonalyze/db"
)

type sacctSummary struct {
	id           uint32
	main         *db.SacctInfo
	steps        []*db.SacctInfo
	maxrss       uint32
	requestedCpu uint64
	usedCpu      uint64
}

func (sc *SacctCommand) Sacct(_ io.Reader, stdout, stderr io.Writer) error {
	var theLog db.SacctCluster
	var err error

	cfg, err := db.MaybeGetConfig(sc.ConfigFile())
	if err != nil {
		return err
	}

	if len(sc.LogFiles) > 0 {
		theLog, err = db.OpenTransientSacctCluster(sc.LogFiles, cfg)
	} else {
		theLog, err = db.OpenPersistentCluster(sc.DataDir, cfg)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store\n%w", err)
	}

	// Read the raw sacct data.

	recordBlobs, dropped, err := theLog.ReadSacctData(
		sc.FromDate,
		sc.ToDate,
		sc.Verbose,
	)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	// TODO: The catenation is expedient, we should be looping over the nested set (or in the
	// future, using an iterator).
	records := uslices.Catenate(recordBlobs)
	if sc.Verbose {
		Log.Infof("%d records read + %d dropped", len(records), dropped)
		UstrStats(stderr, false)
	}

	// Deal with redundant records: due to overlap in data collection runs, a record may be the same
	// as a record we already have.  The pair (JobID, JobStep) is always unique.

	{
		type jobkey struct {
			JobID   uint32
			JobStep Ustr
		}

		recordFilter := make(map[jobkey]bool)
		var rs []*db.SacctInfo
		for _, r := range records {
			var key = jobkey{r.JobID, r.JobStep}
			if !recordFilter[key] {
				recordFilter[key] = true
				rs = append(rs, r)
			}
		}
		if sc.Verbose {
			Log.Infof("%d duplicate records dropped", len(records)-len(rs))
		}
		records = rs
	}

	// Group by job ID

	if sc.Verbose {
		Log.Infof("Working with %d records", len(records))
	}

	byjob := make(map[uint32]*sacctSummary)
	cause := make(map[Ustr]int)
	var mainLess int
	for _, r := range records {
		// TODO: This basically does not allow the "main" line to show up after other lines, is that
		// sensible?  Since we can't have a job without a main line, we will not create an entry
		// here if main is nil.  If we change the logic here to allow main to be added later, we
		// must add post-filtering to remove entries with a main that remains nil at the end.
		if info, ok := byjob[r.JobID]; ok {
			info.steps = append(info.steps, r)
		} else {
			var main *db.SacctInfo
			var steps []*db.SacctInfo
			if r.User != UstrEmpty {
				main = r
			} else {
				steps = []*db.SacctInfo{r}
			}
			if main != nil {
				byjob[r.JobID] = &sacctSummary{
					id:    r.JobID,
					main:  main,
					steps: steps,
				}
			} else {
				mainLess++
			}
		}
		if r.User != UstrEmpty {
			cause[r.State]++
		}
	}
	if sc.Verbose && mainLess > 0 {
		// See above.  This does not happen often and should be the result of the main record just
		// falling on the wrong side of some time quantum (weird) or having been dropped in transit
		// (plausible).
		Log.Infof("%d jobs dropped due to no main record present", mainLess)
	}

	if sc.Verbose {
		Log.Infof("%d jobs", len(byjob))
		for k, v := range cause {
			Log.Infof("  %s: %d", k.String(), v)
		}
	}

	// Filter jobs in byjob on manifest attributes.
	//
	// TODO:
	// - even better filtering on from/to?  The filtering that's happened so far is within the
	//   data store and depending on how that was done there may be some chaff.

	toDelete := make(map[uint32]bool, 0)
	var prior int
	if len(sc.Host) > 0 {
		prior = len(toDelete)
		includeHosts, err := hostglob.NewGlobber(true, sc.Host)
		if err != nil {
			return err
		}
		// We delete a job if none of its nodes match the globber.  It feels like this is a pretty
		// expensive test, the cost is more or less the product of the number of patterns/nodes in
		// the globber and the number of nodes in the expanded nodelist - per job!  Esp the nodelist
		// can be large; the globber will tend to have a very short list (if the query is
		// constructed by a human).  There's a risk here for a DOS, but at least the longer input -
		// the nodelist - is from a controlled source.
		//
		// Possibly caching the (pattern, ExpandPattern(pattern)) pair is worthwhile, but I'd want
		// to see some evidence.
	Outer:
		for id, r := range byjob {
			patterns, err := hostglob.SplitMultiPattern(r.main.NodeList.String())
			if err != nil {
				// Ignore the error here because it is in the input
				break
			}
			for _, pattern := range patterns {
				nodes, err := hostglob.ExpandPattern(pattern)
				if err != nil {
					// Ditto
					continue
				}
				for _, node := range nodes {
					if includeHosts.Match(node) {
						continue Outer
					}
				}
			}
			toDelete[id] = true
		}
		if sc.Verbose {
			Log.Infof("%d filtered by host filter %s", len(toDelete)-prior, sc.Host)
		}
	}
	if len(sc.State) > 0 {
		prior = len(toDelete)
		states := make(map[Ustr]bool)
		for _, k := range sc.State {
			states[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !states[r.main.State] {
				toDelete[id] = true
			}
		}
		if sc.Verbose {
			Log.Infof("%d filtered by state filter %s", len(toDelete)-prior, sc.State)
		}
	}
	if len(sc.User) > 0 {
		prior = len(toDelete)
		users := make(map[Ustr]bool)
		for _, k := range sc.User {
			users[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !users[r.main.User] {
				toDelete[id] = true
			}
		}
		if sc.Verbose {
			Log.Infof("%d filtered by user filter %s", len(toDelete)-prior, sc.User)
		}
	}
	if len(sc.Account) > 0 {
		prior = len(toDelete)
		accounts := make(map[Ustr]bool)
		for _, k := range sc.Account {
			accounts[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !accounts[r.main.Account] {
				toDelete[id] = true
			}
		}
		if sc.Verbose {
			Log.Infof("%d filtered by account filter %s", len(toDelete)-prior, sc.Account)
		}
	}
	if len(sc.Partition) > 0 {
		prior = len(toDelete)
		partitions := make(map[Ustr]bool)
		for _, k := range sc.Partition {
			partitions[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !partitions[r.main.Partition] {
				toDelete[id] = true
			}
		}
		if sc.Verbose {
			Log.Infof("%d filtered by partition filter %s", len(toDelete)-prior, sc.Partition)
		}
	}
	if len(sc.Job) > 0 {
		prior = len(toDelete)
		jobs := make(map[uint32]bool)
		for _, k := range sc.Job {
			jobs[k] = true
		}
		// Slightly subtle (and evolving): if we ask for the overarching ID of an array or het job
		// then we'll find all the parts of it; but we can also ask for an individual part job
		// within the overarching job, and then we'll see only that.
		for id, r := range byjob {
			if !jobs[r.main.JobID] && !jobs[r.main.ArrayJobID] && !jobs[r.main.HetJobID] {
				toDelete[id] = true
			}
		}
		if sc.Verbose {
			Log.Infof("%d filtered by job filter %s", len(toDelete)-prior, sc.Job)
		}
	}
	if sc.SomeGPU || sc.NoGPU {
		prior = len(toDelete)
		for id, r := range byjob {
			if sc.SomeGPU {
				if r.main.ReqGPUS == UstrEmpty {
					toDelete[id] = true
				}
			} else {
				if r.main.ReqGPUS != UstrEmpty {
					toDelete[id] = true
				}
			}
		}
		if sc.Verbose {
			Log.Infof("%d filtered by gpu filter SomeGPU=%v NoGPU=%v",
				len(toDelete)-prior, sc.SomeGPU, sc.NoGPU)
		}
	}
	prior = len(toDelete)
	for id, r := range byjob {
		switch {
		case int64(r.main.ElapsedRaw) < sc.MinRuntime:
			toDelete[id] = true
		case int64(r.main.ElapsedRaw) > sc.MaxRuntime:
			toDelete[id] = true
		}
	}
	if sc.Verbose {
		Log.Infof("%d filtered by elapsed time (runtime) filter: min=%d max=%d",
			len(toDelete)-prior, sc.MinRuntime, sc.MaxRuntime)
	}

	for k := range toDelete {
		delete(byjob, k)
	}

	if sc.Verbose {
		Log.Infof("After filtering: %d jobs.", len(byjob))
	}

	// Partition by job type

	regular := make([]*sacctSummary, 0)
	arrays := make(map[uint32][]*sacctSummary)
	het := make([]*sacctSummary, 0)
	for _, j := range byjob {
		switch {
		case j.main == nil:
			panic("Should have been filtered earlier")
		case j.main.ArrayJobID != 0:
			arrays[j.main.ArrayJobID] = append(arrays[j.main.ArrayJobID], j)
		case j.main.HetJobID != 0:
			het = append(het, j)
		default:
			regular = append(regular, j)
		}
	}

	switch {
	case sc.Array:
		sc.sacctArrayJobs(stdout, arrays)
	case sc.Het:
		panic("Het job output not implemented")
	case sc.Regular:
		sc.sacctRegularJobs(stdout, regular)
	default:
		panic("Unexpected")
	}

	return nil
}

func (sc *SacctCommand) sacctRegularJobs(stdout io.Writer, regular []*sacctSummary) {

	// Compute auxiliary fields we may need during printing

	for _, j := range regular {
		maxmem := j.main.MaxRSS
		for _, s := range j.steps {
			maxmem = max(maxmem, s.MaxRSS)
		}
		j.maxrss = maxmem
		j.requestedCpu =
			uint64(j.main.ReqCPUS) * uint64(j.main.ReqNodes) * uint64(j.main.ElapsedRaw)
		j.usedCpu = j.main.UserCPU + j.main.SystemCPU
	}

	// More filtering

	var toosmall, toobig int
	var toofeeble, toobeefy int
	{
		r := make([]*sacctSummary, 0)
		for _, j := range regular {
			switch {
			case j.main.ReqMem < uint32(sc.MinReservedMem):
				toosmall++
			case j.main.ReqMem > uint32(sc.MaxReservedMem):
				toobig++
			case j.main.ReqCPUS*j.main.ReqNodes < uint32(sc.MinReservedCores):
				toofeeble++
			case j.main.ReqCPUS*j.main.ReqNodes > uint32(sc.MaxReservedCores):
				toobeefy++
			default:
				r = append(r, j)
			}
		}
		regular = r
	}

	if sc.Verbose {
		Log.Infof("regular jobs elided b/c: too small %d, too big %d, too feeble %d, too beefy %d",
			toosmall, toobig, toofeeble, toobeefy)
	}

	if sc.Verbose {
		Log.Infof("After final filtering: %d jobs.", len(regular))
	}

	sc.printRegularJobs(stdout, regular)
}

func (sc *SacctCommand) sacctArrayJobs(stdout io.Writer, arrays map[uint32][]*sacctSummary) {
	// For the array jobs it could look like we get a number of "elements" that corresponds to the
	// number of concurrent array jobs?  But that could really be a result of incomplete input data
	// at this point.
	//
	// Notably the index values are completely controllable by the application, there is no
	// guarantee that the range is dense or has particular start/stop values, except that they are
	// nonnegative.

	// For an array job, how do we view it in terms of resource allocation?  Do all elements get the
	// same parameters?  Do we assess them individually, or use some sort of max across the elements
	// or the steps in the elements?

	// id is ArrayJobId
	for id, js := range arrays {
		fmt.Fprintf(
			stdout,
			"\n%d array %s %d elements %s\n",
			id, js[0].main.User, len(js), js[0].main.State)
		sc.sacctRegularJobs(stdout, js)
		/*
		for _, j := range js {
			fmt.Fprintf(
				stdout,
				"  index=%d, id=%d %d steps\n",
				j.main.ArrayIndex, j.main.JobID, len(j.steps))
			for _, s := range j.steps {
				fmt.Fprintf(stdout, "    step %s %s\n", s.JobStep, s.State)
			}
		}
		*/
	}
}
