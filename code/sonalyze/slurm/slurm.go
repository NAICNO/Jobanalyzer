/*
   let's start with a canned query to drive exporation of data and postprocessing.

   -all-atleast key=n,...

   -any-atmost key=n,...
*/

/*

   suppose we want to find jobs that used less memory than they requested.

   we're interested in jobs that completed in a time window, so --from Nd as usual.

   other record and job filtering options can clearly apply in the normal way.

   fundamentally a query about average memory use?  ReqMem vs MaxRSS, ReqMem vs AvgRSS.  We could
   dump every job and postprocess in awk, but following perhaps `-rss-avg-below 50` and
   `-rss-max-below 50` would be good starting points. (`or` can be expressed as multiple runs and
   `and` is implicit if both switches are enabled.)

   Ditto, `-cpu-avg-below 50` queries the relationship (user+system)/cputime where the latter
   is just cores*elapsed.  (The cputimeraw field is really completely redundant.)

   Ditto, `-realtime-below 50` if there is a time limit, although this does not seem important -
   once the job is done, something else will be scheduled.  At most this impacts the ability to
   create an optimal schedule.

   This gives us a list of job numbers (or job data in geneal).  now it will be possible to use
   sonar data to look at what the job did.  Suppose we are querying memory usage.  the memory
   profile for the job over time is a signal can be used to characterize the job.

   For the management report we may wish to have what, exactly?  What is actionable information?

   Suppose a "good" job is one that has peak memory use > 75% of its reservation and average cpu use
   > 75% of its reservation (elapsed*cores).  we'll discover immediately that the autotekst jobs
   have requested 8 cpus but use only one (at about 100%).  this is the kind of thing we should be
   able to discover with a profile.  (unless there is a "smallest reservation" at 8 cpus.)

   We could have -above-all cpu=75,mem=75 and -below-any cpu=75,mem=75 to list good and bad jobs,
   where above means ">=" and below means "<".

   so i think we'll have

     sonalyze slurm-jobs <standard-options> <filter-options>

   very much like `sonalyze jobs`, with different filtering options (a little goes a long way).

   we don't have samples so the record options might not apply.

   Some of the RecordFilterArgs have analogues here but filter jobs, not records, so we'll
   want to add those.  At least these:

   - user
   - job

   (The -exclude-user filter has paid for itself with sonar data but that's because of all
   the system users, we won't have that issue here I think.)

   ----

   Rough priority order:

   Set up logging on Fox and ingest on naic-monitor to generate the data while we're working out
   the code.

   => After reading records, they must be postprocessed to make sense of them probably.  I still
      don't know how CPU time / memory for individual steps are accounted, and how we aggregate
      these.

      Looking at the data for the array job in my homedir on Fox, it could look like all user CPU
      time is aggregated up into the top-line job by Slurm, so there may not be any real aggregation
      to be done, just selecting that.

      It's possible that "Aggregation" is creating a tree or cluster of items that belong together,
      for when we need that, but that for the task at hand we may only need the top-line item?

   Is there a slurm-parse verb too?

 */

package slurm

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"strings"
	"time"

	//. "sonalyze/common"
	. "sonalyze/command"
	_ "sonalyze/db"
)

const (
	defaultMinElapsed = 600
	defaultMinReservedMem = 10
	defaultMinReservedCores = 4
)

type SacctCommand struct /* implements AnalysisCommand */ {
	// Almost SharedArgs, but HostArgs instead of RecordFilterArgs
	// TODO: add -user, -job
	DevArgs
	SourceArgs
	HostArgs
	VerboseArgs
	ConfigFileArgs

	// Selections
	// COMPLETED, TIMEOUT etc - "or" these, except "" for all
	State []string
	// User names
	User []string
	// Account names
	Account []string
	MinElapsed uint
	MinReservedMem uint
	MinReservedCores uint

	// Printing
	Fmt string

	// Synthesized and other
	printFields []string
	printOpts   *FormatOptions
}

var _ = AnalysisCommand((*SacctCommand)(nil))

func (_ *SacctCommand) Summary() []string {
	return []string{
		"Extract information from slurm data independent of sample data",
	}
}

func (sc *SacctCommand) Add(fs *flag.FlagSet) {
	sc.DevArgs.Add(fs)
	sc.SourceArgs.Add(fs)
	sc.HostArgs.Add(fs)
	sc.VerboseArgs.Add(fs)
	sc.ConfigFileArgs.Add(fs)
	fs.UintVar(&sc.MinElapsed, "min-elapsed", defaultMinElapsed,
		"Select jobs with elapsed time at least this")
	fs.UintVar(&sc.MinReservedMem, "min-reserved-mem", defaultMinReservedMem,
		"Select jobs with reserved memory at least this (GB)")
	fs.UintVar(&sc.MinReservedCores, "min-reserved-cores", defaultMinReservedCores,
		"Select jobs with reserved cores (cpus * nodes) at least this")
	fs.Var(NewRepeatableString(&sc.State), "state",
		"Select jobs with state `state,...`: COMPLETED, CANCELLED, DEADLINE, FAILED, OUT_OF_MEMORY, TIMEOUT")
	fs.Var(NewRepeatableString(&sc.User), "user",
		"Select jobs with user `user1,...`")
	fs.Var(NewRepeatableString(&sc.Account), "account",
		"Select jobs with account `account1,...`")
	fs.StringVar(&sc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (sc *SacctCommand) ReifyForRemote(x *Reifier) error {
	x.String("fmt", sc.Fmt)
	x.RepeatableString("state", sc.State)
	x.RepeatableString("user", sc.User)
	x.RepeatableString("account", sc.Account)
	x.Uint("min-elapsed", sc.MinElapsed)
	return errors.Join(
		sc.DevArgs.ReifyForRemote(x),
		sc.SourceArgs.ReifyForRemote(x),
		sc.HostArgs.ReifyForRemote(x),
		sc.ConfigFileArgs.ReifyForRemote(x),
	)
}

func (sc *SacctCommand) Validate() error {
	var e1, e2, e3, e4, e5, e6 error
	e1 = sc.DevArgs.Validate()
	e2 = sc.SourceArgs.Validate()
	e3 = sc.HostArgs.Validate()
	e4 = sc.VerboseArgs.Validate()
	e5 = sc.ConfigFileArgs.Validate()

	var others map[string]bool
	sc.printFields, others, e6 = ParseFormatSpec(sacctDefaultFields, sc.Fmt, sacctFormatters, sacctAliases)
	if e6 == nil && len(sc.printFields) == 0 {
		e6 = errors.New("No output fields were selected in format string")
	}
	sc.printOpts = StandardFormatOptions(others, DefaultFixed)

	return errors.Join(e1, e2, e3, e4, e5, e6)
}

func (sc *SacctCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(sc.Fmt, sacctHelp, sacctFormatters, sacctAliases, sacctDefaultFields)
}

const sacctHelp = `
parse
  Aggregate Slurm sacct data into data about jobs and present them.
`

const sacctDefaultFields = "JobID,JobStep,JobName,User,Account,rcpu,rmem"

// MT: Constant after initialization; immutable
var sacctAliases = map[string][]string{
	"default": strings.Split(sacctDefaultFields, ","),
}

type sacctCtx = bool			// fixed format or not

// MT: Constant after initialization; immutable
var sacctFormatters = map[string]Formatter[*sacctSummary, sacctCtx]{
	"Timestamp": {
		func (d *sacctSummary, _ sacctCtx) string {
			return time.Unix(d.main.Timestamp, 0).UTC().Format(time.RFC3339)
		},
		"Time record was obtained",
	},
	"Start": {
		func (d *sacctSummary, _ sacctCtx) string {
			if d.main.Start == 0 {
				return "Unknown"
			}
			return time.Unix(d.main.Start, 0).UTC().Format(time.RFC3339)
		},
		"Start time of job, if any",
	},
	"End": {
		func (d *sacctSummary, _ sacctCtx) string {
			return time.Unix(d.main.End, 0).UTC().Format(time.RFC3339)
		},
		"End time of job",
	},
	"Submit": {
		func (d *sacctSummary, _ sacctCtx) string {
			return time.Unix(d.main.Submit, 0).UTC().Format(time.RFC3339)
		},
		"Submit time of job",
	},
	"RequestedCPU": {
		func (d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.requestedCpu)
		},
		"Requested CPU time (elapsed * cores * nodes)",
	},
	"UsedCPU": {
		func (d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.usedCpu)
		},
		"Used CPU time",
	},
	"rcpu": {
		func (d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(math.Round(100 * float64(d.usedCpu) / float64(d.requestedCpu)))
		},
		"Percent cpu utilization",
	},
	"rmem": {
		func (d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(math.Round(100 * float64(d.maxrss) / float64(d.main.ReqMem)))
		},
		"Percent memory utilization",
	},
	"User": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.User.String()
		},
		"Job's user",
	},
	"JobName": {
		func(d *sacctSummary, ctx sacctCtx) string {
			s := d.main.JobName.String()
			if ctx && len(s) > 30 {
				s = s[:30]
			}
			return s
		},
		"Job name",
	},
	"State": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.State.String()
		},
		"Job completion state",
	},
	"Account": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.Account.String()
		},
		"Job's account",
	},
	"Reservation": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.Reservation.String()
		},
		"Job's reservation, if any",
	},
	"NodeList": {
		func(d *sacctSummary, _ sacctCtx) string {
			return d.main.NodeList.String()
		},
		"Job's node list",
	},
	"JobID": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.JobID)
		},
		"Primary Job ID",
	},
	"MaxRSS": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.maxrss)
		},
		"Max resident set size (RSS) across all steps",
	},
	"RequestedMem": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ReqMem)
		},
		"Requested memory",
	},
	"Suspended": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.Suspended)
		},
		"Time suspended",
	},
	"Timelimit": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.TimelimitRaw)
		},
		"Time limit in seconds",
	},
	"ExitCode": {
		func(d *sacctSummary, _ sacctCtx) string {
			return fmt.Sprint(d.main.ExitCode)
		},
		"Exit code",
	},
}
