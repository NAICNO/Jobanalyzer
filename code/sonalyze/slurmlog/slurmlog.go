package slurmlog

import (
	"fmt"
	"strings"
	"time"

	"go-utils/hostglob"
	umaps "go-utils/maps"
	uslices "go-utils/slices"

	. "sonalyze/common"
	"sonalyze/db"
)

type QueryFilter struct {
	Host        []string
	State       []string // COMPLETED, TIMEOUT etc - "or" these, except "" for all
	User        []string
	Account     []string
	Partition   []string
	Reservation []string
	GpuType     []string // Prefixes
	Job         []uint32
	SomeGPU     bool
	NoGPU       bool
	MinRuntime  int64
	MaxRuntime  int64 // 0 means "not set"
}

type SlurmJob struct {
	Id    uint32
	Main  *db.SacctInfo
	Steps []*db.SacctInfo
}

func Query(
	theLog db.SacctCluster,
	fromDate, toDate time.Time,
	filter QueryFilter,
	verbose bool,
) ([]*SlurmJob, error) {
	// Read the raw sacct data.

	recordBlobs, dropped, err := theLog.ReadSacctData(
		fromDate,
		toDate,
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	// TODO: The catenation is expedient, we should be looping over the nested set (or in the
	// future, using an iterator).
	records := uslices.Catenate(recordBlobs)
	if verbose {
		Log.Infof("%d records read + %d dropped", len(records), dropped)
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
		if verbose {
			Log.Infof("%d duplicate records dropped", len(records)-len(rs))
		}
		records = rs
	}

	// Group by job ID

	if verbose {
		Log.Infof("Working with %d records", len(records))
	}

	byjob := make(map[uint32]*SlurmJob)
	cause := make(map[Ustr]int)
	var mainLess int
	for _, r := range records {
		// TODO: This basically does not allow the "main" line to show up after other lines, is that
		// sensible?  Since we can't have a job without a main line, we will not create an entry
		// here if main is nil.  If we change the logic here to allow main to be added later, we
		// must add post-filtering to remove entries with a main that remains nil at the end.
		if info, ok := byjob[r.JobID]; ok {
			info.Steps = append(info.Steps, r)
		} else {
			var main *db.SacctInfo
			var steps []*db.SacctInfo
			if r.User != UstrEmpty {
				main = r
			} else {
				steps = []*db.SacctInfo{r}
			}
			if main != nil {
				byjob[r.JobID] = &SlurmJob{
					Id:    r.JobID,
					Main:  main,
					Steps: steps,
				}
			} else {
				mainLess++
			}
		}
		if r.User != UstrEmpty {
			cause[r.State]++
		}
	}
	if verbose && mainLess > 0 {
		// See above.  This does not happen often and should be the result of the main record just
		// falling on the wrong side of some time quantum (weird) or having been dropped in transit
		// (plausible).
		Log.Infof("%d jobs dropped due to no main record present", mainLess)
	}

	if verbose {
		Log.Infof("%d jobs", len(byjob))
		for k, v := range cause {
			Log.Infof("  %s: %d", k.String(), v)
		}
	}

	err = filterJobs(byjob, filter, verbose)
	if err != nil {
		return nil, err
	}

	return umaps.Values(byjob), nil
}

func FilterJobs(
	jobs []*SlurmJob,
	filter QueryFilter,
	verbose bool,
) ([]*SlurmJob, error) {
	byjob := make(map[uint32]*SlurmJob)
	for _, j := range jobs {
		byjob[j.Id] = j
	}
	err := filterJobs(byjob, filter, verbose)
	if err != nil {
		return nil, err
	}
	return umaps.Values(byjob), nil
}

func filterJobs(byjob map[uint32]*SlurmJob, filter QueryFilter, verbose bool) error {
	// Filter jobs in byjob on manifest attributes.
	//
	// TODO:
	// - even better filtering on from/to?  The filtering that's happened so far is within the
	//   data store and depending on how that was done there may be some chaff.

	toDelete := make(map[uint32]bool, 0)
	var prior int
	if len(filter.Host) > 0 {
		prior = len(toDelete)
		hosts, err := NewHosts(true, filter.Host)
		if err != nil {
			return err
		}
		includeHosts := hosts.HostnameGlobber()
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
			patterns, err := hostglob.SplitMultiPattern(r.Main.NodeList.String())
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
		if verbose {
			Log.Infof("%d filtered by host filter %s", len(toDelete)-prior, filter.Host)
		}
	}
	filterByString(
		byjob, toDelete, verbose,
		"state",
		filter.State,
		func(j *SlurmJob) Ustr { return j.Main.State },
	)
	filterByString(
		byjob, toDelete, verbose,
		"user",
		filter.User,
		func(j *SlurmJob) Ustr { return j.Main.User },
	)
	filterByString(
		byjob, toDelete, verbose,
		"account",
		filter.Account,
		func(j *SlurmJob) Ustr { return j.Main.Account },
	)
	filterByString(
		byjob, toDelete, verbose,
		"partition",
		filter.Partition,
		func(j *SlurmJob) Ustr { return j.Main.Partition },
	)
	filterByString(
		byjob, toDelete, verbose,
		"reservation",
		filter.Reservation,
		func(j *SlurmJob) Ustr { return j.Main.Reservation },
	)

	// For GpuType, we could build a regexp of the requested GPU types or we could just run
	// HasPrefix on every record for every type.  Mostly there will be zero or one Gpu types,
	// so go with the latter for now.

	if len(filter.GpuType) > 0 {
		prior := len(toDelete)
		for id, r := range byjob {
			if r.Main.ReqGPUS != UstrEmpty {
				request := r.Main.ReqGPUS.String()
				match := false
				for _, f := range filter.GpuType {
					if strings.HasPrefix(request, f) {
						match = true
					}
				}
				if !match {
					toDelete[id] = true
				}
			} else {
				toDelete[id] = true
			}
		}
		if verbose {
			Log.Infof("%d filtered by gpu type filter %v", len(toDelete)-prior, filter.GpuType)
		}
	}

	if len(filter.Job) > 0 {
		prior = len(toDelete)
		jobs := make(map[uint32]bool)
		for _, k := range filter.Job {
			jobs[k] = true
		}
		// Slightly subtle (and evolving): if we ask for the overarching ID of an array or het job
		// then we'll find all the parts of it; but we can also ask for an individual part job
		// within the overarching job, and then we'll see only that.
		for id, r := range byjob {
			if !jobs[r.Main.JobID] && !jobs[r.Main.ArrayJobID] && !jobs[r.Main.HetJobID] {
				toDelete[id] = true
			}
		}
		if verbose {
			Log.Infof("%d filtered by job filter %s", len(toDelete)-prior, filter.Job)
		}
	}
	if filter.SomeGPU || filter.NoGPU {
		prior = len(toDelete)
		for id, r := range byjob {
			if filter.SomeGPU {
				if r.Main.ReqGPUS == UstrEmpty {
					toDelete[id] = true
				}
			} else {
				if r.Main.ReqGPUS != UstrEmpty {
					toDelete[id] = true
				}
			}
		}
		if verbose {
			Log.Infof("%d filtered by gpu filter SomeGPU=%v NoGPU=%v",
				len(toDelete)-prior, filter.SomeGPU, filter.NoGPU)
		}
	}
	prior = len(toDelete)
	for id, r := range byjob {
		switch {
		case int64(r.Main.ElapsedRaw) < filter.MinRuntime:
			toDelete[id] = true
		case filter.MaxRuntime > 0 && int64(r.Main.ElapsedRaw) > filter.MaxRuntime:
			toDelete[id] = true
		}
	}
	if verbose {
		Log.Infof("%d filtered by elapsed time (runtime) filter: min=%d max=%d",
			len(toDelete)-prior, filter.MinRuntime, filter.MaxRuntime)
	}

	for k := range toDelete {
		delete(byjob, k)
	}

	if verbose {
		Log.Infof("After filtering: %d jobs.", len(byjob))
	}

	return nil
}

func filterByString(
	byjob map[uint32]*SlurmJob,
	toDelete map[uint32]bool,
	verbose bool,
	what string,
	filters []string,
	get func(j *SlurmJob) Ustr,
) {
	if len(filters) > 0 {
		prior := len(toDelete)
		fs := make(map[Ustr]bool)
		for _, k := range filters {
			fs[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !fs[get(r)] {
				toDelete[id] = true
			}
		}
		if verbose {
			Log.Infof("%d filtered by %s filter %v", len(toDelete)-prior, what, filters)
		}
	}
}
