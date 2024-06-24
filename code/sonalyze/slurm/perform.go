// It seems pretty clear that:
//  - the command should be `sacct-jobs` and not just `sacct`
//  - we'll want to implement filtering, basic aggregation, and printing/formatting in the standard way
//  - any reports should be written on the outside
//  - additional filtering
//    - host/node glob
//    - job type (regular, array)
//    - job name positive/negative glob

package slurm

import (
	_ "cmp"
	"fmt"
	"io"
	_ "math"
	_ "slices"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
)

type sacctSummary struct {
	id uint32
	main *db.SacctInfo
	steps []*db.SacctInfo
	maxrss uint32
	requestedCpu uint64
	usedCpu uint64
}

func (sc *SacctCommand) Sacct(_ io.Reader, stdout, stderr io.Writer) error {
	var theLog db.SacctCluster
	var err error

	cfg, err := MaybeGetConfig(sc.ConfigFile())
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
	if err != nil {
		return fmt.Errorf("Failed to open log store\n%w", err)
	}

	// Read the raw sacct data.

	records, dropped, err := theLog.ReadSacctData(
		sc.FromDate,
		sc.ToDate,
		sc.Verbose,
	)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	if sc.Verbose {
		Log.Infof("%d records read + %d dropped", len(records), dropped)
		UstrStats(stderr, false)
	}

	// Deal with redundant records: due to overlap in data collection runs, a record may be the same
	// as a record we already have.  The pair (JobID, JobStep) is always unique.

	{
		type jobkey struct {
			JobID uint32
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
		records = rs
	}

	// Group by job ID

	byjob := make(map[uint32]*sacctSummary)
	cause := make(map[Ustr]int)
	for _, r := range records {
		if info, ok := byjob[r.JobID]; ok {
			info.steps = append(info.steps, r)
		} else {
			var main *db.SacctInfo
			var steps []*db.SacctInfo
			if r.User != UstrEmpty {
				main = r
			} else {
				steps = []*db.SacctInfo{ r }
			}
			byjob[r.JobID] = &sacctSummary{
				id: r.JobID,
				main: main,
				steps: steps,
			}
		}
		if r.User != UstrEmpty {
			cause[r.State]++
		}
	}

	if sc.Verbose {
		Log.Infof("%d jobs.\n", len(byjob))
		for k, v := range cause {
			Log.Infof("  %s: %d\n", k.String(), v)
		}
	}

	// Filter jobs in byjob on manifest attributes.
	//
	// TODO:
	// - host args, useful for filtering for eg gpu nodes
	// - job id

	toDelete := make([]uint32, 0)
	if len(sc.State) > 0 {
		states := make(map[Ustr]bool)
		for _, k := range sc.State {
			states[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !states[r.main.State] {
				toDelete = append(toDelete, id)
			}
		}
	}
	if len(sc.User) > 0 {
		users := make(map[Ustr]bool)
		for _, k := range sc.User {
			users[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !users[r.main.User] {
				toDelete = append(toDelete, id)
			}
		}
	}
	if len(sc.Account) > 0 {
		accounts := make(map[Ustr]bool)
		for _, k := range sc.Account {
			accounts[StringToUstr(k)] = true
		}
		for id, r := range byjob {
			if !accounts[r.main.Account] {
				toDelete = append(toDelete, id)
			}
		}
	}
	for _, k := range toDelete {
		delete(byjob, k)
	}

	// Partition by job type

	regular := make([]*sacctSummary, 0)
	arrays := make(map[uint32][]*sacctSummary)
	het := make([]*sacctSummary, 0)
	broken := make([]*sacctSummary, 0)
	for _, j := range byjob {
		switch {
		case j.main == nil:
			broken = append(broken, j)
		case j.main.ArrayJobID != 0:
			arrays[j.main.ArrayJobID] = append(arrays[j.main.ArrayJobID], j)
		case j.main.HetJobID != 0:
			het = append(het, j)
		default:
			regular = append(regular, j)
		}
	}

	sc.sacctRegularJobs(stdout, regular)
	//sc.sacctArrayJobs(stdout, arrays)

	return nil
}

func (sc *SacctCommand) sacctRegularJobs(stdout io.Writer, regular []*sacctSummary) {

	// Compute auxiliary fields we may need

	for _, j := range regular {
		maxmem := j.main.MaxRSS
		for _, s := range j.steps {
			maxmem = max(maxmem, s.MaxRSS)
		}
		j.maxrss = maxmem
		j.requestedCpu = uint64(j.main.ReqCPUS) * uint64(j.main.ReqNodes) * uint64(j.main.ElapsedRaw)
		j.usedCpu = j.main.UserCPU + j.main.SystemCPU
	}

	// More filtering

	var tooshort int
	var toosmall int
	var toofeeble int
	{
		r := make([]*sacctSummary, 0)
		for _, j := range regular {
			switch {
			case j.main.ElapsedRaw < uint32(sc.MinElapsed):
				tooshort++
			case j.main.ReqMem < uint32(sc.MinReservedMem):
				toosmall++
			case j.main.ReqCPUS * j.main.ReqNodes < uint32(sc.MinReservedCores):
				toofeeble++
			default:
				r = append(r, j)
			}
		}
		regular = r
	}

	if sc.Verbose {
		Log.Infof("\nregular jobs elided b/c: too short %d, too small %d, too feeble %d",
			tooshort, toosmall, toofeeble)
	}

	FormatData(stdout, sc.printFields, sacctFormatters, sc.printOpts, regular,
		sc.printOpts.Fixed)

	// Display
	/*
	slices.SortFunc(regular, func(a, b *sacctSummary) int {
		return cmp.Compare(a.id, b.id)
	})

	for _, j := range regular {
		fmt.Fprintf(
			stdout,
			"\n%d regular %s (%s), %d steps, elapsed %ds %s %s\n",
			j.id, j.main.User, j.main.Account, len(j.steps), j.main.ElapsedRaw, j.main.State, j.main.JobName)

		fmt.Fprintf(
			stdout,
			"  memory requested = %dG, peakrss = %dG, ratio = %d%%\n",
			j.main.ReqMem,
			j.maxrss,
			int(math.Round(100*float64(j.maxrss)/float64(j.main.ReqMem))))

		fmt.Fprintf(
			stdout,
			"  cpu requested = %ds, used = %ds, ratio = %d%% (%d cores, %d nodes)\n",
			j.requestedCpu,
			j.usedCpu,
			int(math.Round(100 * float64(j.usedCpu) / float64(j.requestedCpu))),
			j.main.ReqCPUS,
			j.main.ReqNodes,
		)

		if j.main.State.String() == "TIMEOUT" {
			fmt.Fprintf(stdout, "  Time limit %ds\n", j.main.TimelimitRaw)
		}
	}
	*/
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
		for _, j := range js {
			fmt.Fprintf(
				stdout,
				"  index=%d, id=%d %d steps\n",
				j.main.ArrayIndex, j.main.JobID, len(j.steps))
			for _, s := range j.steps {
				fmt.Fprintf(stdout, "    step %s %s\n", s.JobStep, s.State)
			}
		}
	}
}
