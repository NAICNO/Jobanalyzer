package sacct

import (
	"fmt"
	"io"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/slurmlog"
)

type sacctSummary struct {
	*slurmlog.SlurmJob
	maxrss       uint32
	requestedCpu uint64
	usedCpu      uint64
}

func (sc *SacctCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
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
		return fmt.Errorf("Failed to open log store: %v", err)
	}

	jobs, err := slurmlog.Query(
		theLog,
		sc.FromDate,
		sc.ToDate,
		slurmlog.QueryFilter{
			Host:       sc.Host,
			State:      sc.State,
			User:       sc.User,
			Account:    sc.Account,
			Partition:  sc.Partition,
			Job:        sc.Job,
			SomeGPU:    sc.SomeGPU,
			NoGPU:      sc.NoGPU,
			MinRuntime: sc.MinRuntime,
			MaxRuntime: sc.MaxRuntime,
		},
		sc.Verbose,
	)
	if err != nil {
		return err
	}

	// Partition by job type

	regular := make([]*slurmlog.SlurmJob, 0)
	arrays := make(map[uint32][]*slurmlog.SlurmJob)
	het := make([]*slurmlog.SlurmJob, 0)
	for _, j := range jobs {
		switch {
		case j.Main == nil:
			panic("Should have been filtered earlier")
		case j.Main.ArrayJobID != 0:
			arrays[j.Main.ArrayJobID] = append(arrays[j.Main.ArrayJobID], j)
		case j.Main.HetJobID != 0:
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

func (sc *SacctCommand) sacctRegularJobs(stdout io.Writer, regularJobs []*slurmlog.SlurmJob) {

	// Compute auxiliary fields we may need during printing

	regular := make([]*sacctSummary, 0)
	for _, j := range regularJobs {
		maxmem := j.Main.MaxRSS
		for _, s := range j.Steps {
			maxmem = max(maxmem, s.MaxRSS)
		}
		regular = append(regular, &sacctSummary{
			SlurmJob:     j,
			maxrss:       maxmem,
			requestedCpu: uint64(j.Main.ReqCPUS) * uint64(j.Main.ReqNodes) * uint64(j.Main.ElapsedRaw),
			usedCpu:      j.Main.UserCPU + j.Main.SystemCPU,
		})
	}

	// More filtering

	// TODO: Why here?

	var toosmall, toobig int
	var toofeeble, toobeefy int
	{
		r := make([]*sacctSummary, 0)
		for _, j := range regular {
			switch {
			case j.Main.ReqMem < uint32(sc.MinReservedMem):
				toosmall++
			case j.Main.ReqMem > uint32(sc.MaxReservedMem):
				toobig++
			case j.Main.ReqCPUS*j.Main.ReqNodes < uint32(sc.MinReservedCores):
				toofeeble++
			case j.Main.ReqCPUS*j.Main.ReqNodes > uint32(sc.MaxReservedCores):
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

func (sc *SacctCommand) sacctArrayJobs(stdout io.Writer, arrays map[uint32][]*slurmlog.SlurmJob) {
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
			id, js[0].Main.User, len(js), js[0].Main.State)
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
