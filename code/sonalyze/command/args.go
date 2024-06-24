package command

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	. "sonalyze/common"
	"sonalyze/sonarlog"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// DevArgs are for development and their inclusion can be controlled with the devArgs setting,
// below.

type DevArgs struct {
	CpuProfile string
}

const devArgs = true

func (d *DevArgs) CpuProfileFile() string {
	return d.CpuProfile
}

func (d *DevArgs) Add(fs *flag.FlagSet) {
	if devArgs {
		fs.StringVar(&d.CpuProfile, "cpuprofile", "",
			"(Development) write cpu profile to `filename`")
	}
}

func (d *DevArgs) ReifyForRemote(x *Reifier) error {
	if d.CpuProfile != "" {
		return errors.New("-cpuprofile not allowed with remote execution")
	}
	return nil
}

func (d *DevArgs) Validate() error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// You wouldn't think -v would be so complicated.

type VerboseArgs struct {
	Verbose bool
}

func (va *VerboseArgs) Add(fs *flag.FlagSet) {
	fs.BoolVar(&va.Verbose, "v", false, "Print verbose diagnostics to stderr")
	// The Rust version allows both -v and --verbose
	fs.BoolVar(&va.Verbose, "verbose", false, "Print verbose diagnostics to stderr")
}

func (va *VerboseArgs) Validate() error {
	return nil
}

func (va *VerboseArgs) VerboseFlag() bool {
	return va.Verbose
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Handle -data-dir

type DataDirArgs struct {
	DataDir string
}

func (dd *DataDirArgs) Add(fs *flag.FlagSet) {
	fs.StringVar(&dd.DataDir, "data-dir", "",
		"Select the root `directory` for log files [default: $SONAR_ROOT or $HOME/sonar/data]")
	fs.StringVar(&dd.DataDir, "data-path", "", "Alias for -data-dir `directory`")
}

// TODO: $SONAR_ROOT is a completely inappropriate name and $HOME/sonar/data is almost certainly the
// wrong directory, because what we want is a *cluster* directory under some root directory.  And
// the cluster name cannot be defaulted.  So this logic should be rethought or (probably) deleted.
//
// However, we could require that the directory exist, if the argument is not "", see eg the
// `infiltrate` source.  This would be a useful service.

func (dd *DataDirArgs) Validate() error {
	if dd.DataDir != "" {
		dd.DataDir = path.Clean(dd.DataDir)
	} else if d := os.Getenv("SONAR_ROOT"); d != "" {
		dd.DataDir = path.Clean(d)
	} else if d := os.Getenv("HOME"); d != "" {
		dd.DataDir = path.Clean(path.Join(d, "/sonar/data"))
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// RemotingArgs pertain to specifying a remote sonalyze service.  Note that the meaning of the
// -auth-file depends on the operation: for `add` it would normally be `cluster:cluster-password`
// pairs, not `user:password`.

type RemotingArgs struct {
	Remote   string
	Cluster  string
	AuthFile string

	Remoting bool
}

func (ra *RemotingArgs) Add(fs *flag.FlagSet) {
	fs.StringVar(&ra.Remote, "remote", "",
		"Select a remote `url` to serve the query [default: none].  Requires -cluster.")
	fs.StringVar(&ra.Cluster, "cluster", "",
		"Select the cluster `name` for which we want data [default: none].  For use with -remote.")
	fs.StringVar(&ra.AuthFile, "auth-file", "",
		"Provide a `file` with username:password [default: none].  For use with -remote.")
}

func (ra *RemotingArgs) Validate() error {
	if ra.Remote != "" || ra.Cluster != "" {
		if ra.Remote == "" || ra.Cluster == "" {
			return fmt.Errorf("-remote and -cluster must be used together")
		}
		ra.Remoting = true
	}
	return nil
}

func (va *RemotingArgs) RemotingFlags() *RemotingArgs {
	return va
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// SourceArgs pertain to source file location and initial filtering-by-location, though the
// -from/-to arguments are also used to filter records.

type SourceArgs struct {
	DataDirArgs
	RemotingArgs
	HaveFrom bool
	FromDate time.Time
	HaveTo   bool
	ToDate   time.Time
	LogFiles []string

	fromDateStr string
	toDateStr   string
}

func (s *SourceArgs) Add(fs *flag.FlagSet) {
	s.DataDirArgs.Add(fs)
	s.RemotingArgs.Add(fs)
	fs.StringVar(&s.fromDateStr, "from", "",
		"Select records by this `time` and later.  Format can be YYYY-MM-DD, or Nd or Nw\n"+
			"signifying N days or weeks ago [default: 1d, ie 1 day ago]")
	fs.StringVar(&s.fromDateStr, "f", "", "Short for -from `time`")
	fs.StringVar(&s.toDateStr, "to", "",
		"Select records by this `time` and earlier.  Format can be YYYY-MM-DD, or Nd or Nw\n"+
			"signifying N days or weeks ago [default: now]")
	fs.StringVar(&s.toDateStr, "t", "", "Short for -to `time`")
}

func (s *SourceArgs) ReifyForRemote(x *Reifier) error {
	// RemotingArgs don't have ReifyForRemote

	// Validate() has already checked that DataDir, LogFiles, Remote, Cluster, and AuthFile are
	// consistent for remote or local execution; none of those except Cluster is forwarded.
	x.String("cluster", s.Cluster)
	x.String("from", s.fromDateStr)
	x.String("to", s.toDateStr)
	return nil
}

func (s *SourceArgs) SetRestArguments(args []string) {
	s.LogFiles = args
}

func (s *SourceArgs) Validate() error {
	err := s.RemotingArgs.Validate()
	if err != nil {
		return err
	}

	if s.Remoting {
		// If remoting then no local data sources are allowed, so don't compute default data dirs by
		// calling Validate(), it would confuse the matter - just disallow explicit values.  (This
		// is a small abstraction leak.)
		if s.DataDir != "" {
			return fmt.Errorf("-data-dir may not be used with -remote or -cluster")
		}
		if len(s.LogFiles) > 0 {
			return fmt.Errorf("-- logfile ... may not be used with -remote or -cluster")
		}
	} else {
		// Compute and clean the dataDir and clean any logfiles.  If we have neither logfiles nor
		// dataDir then signal an error.
		err := s.DataDirArgs.Validate()
		if err != nil {
			return err
		}
		if len(s.LogFiles) > 0 {
			for i := 0; i < len(s.LogFiles); i++ {
				s.LogFiles[i] = path.Clean(s.LogFiles[i])
			}
		} else if s.DataDir == "" {
			return fmt.Errorf("Required -data-dir or -- logfile ...")
		}
	}

	// The song and dance with `HaveFrom` and `HaveTo` is this: when a list of files is present then
	// `-from` and `-to` are inferred from the file contents, so long as they are not present on the
	// command line.

	now := time.Now().UTC()
	if s.fromDateStr != "" {
		var err error
		s.FromDate, err = ParseRelativeDateUtc(now, s.fromDateStr, false)
		if err != nil {
			return fmt.Errorf("Invalid -from argument %s", s.fromDateStr)
		}
		s.HaveFrom = true
	} else {
		s.FromDate = now.AddDate(0, 0, -1)
		s.HaveFrom = len(s.LogFiles) == 0
	}

	if s.toDateStr != "" {
		var err error
		s.ToDate, err = ParseRelativeDateUtc(now, s.toDateStr, true)
		if err != nil {
			return fmt.Errorf("Invalid -to argument %s", s.toDateStr)
		}
		s.HaveTo = true
	} else {
		s.ToDate = now
		s.HaveFrom = len(s.LogFiles) == 0
	}

	if s.FromDate.After(s.ToDate) {
		return errors.New("The -from time is greater than the -to time")
	}

	return nil
}

// Grab FromDate and ToDate from args if available, otherwise infer from the bounds, otherwise use
// the defaults.  Return as int64 timestamps compatible with the Sample timestamps.

func (args *SourceArgs) InterpretFromToWithBounds(bounds sonarlog.Timebounds) (int64, int64) {
	var from, to int64
	if args.HaveFrom || len(bounds) == 0 {
		from = args.FromDate.Unix()
	} else {
		from = math.MaxInt64
		for _, v := range bounds {
			from = min(from, v.Earliest)
		}
	}
	if args.HaveTo || len(bounds) == 0 {
		to = args.ToDate.Unix()
	} else {
		for _, v := range bounds {
			to = max(to, v.Latest)
		}
	}
	return from, to
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// HostArgs and RecordFilterArgs pertain to including and excluding records by purely record-local
// criteria.  In addition to these, the -from / -to arguments of the SourceArgs are applied as
// record filters.

type HostArgs struct {
	Host []string
}

func (h *HostArgs) Add(fs *flag.FlagSet) {
	fs.Var(NewRepeatableString(&h.Host), "host",
		"Select records for this `host` (repeatable) [default: all]")
}

func (h *HostArgs) ReifyForRemote(x *Reifier) error {
	x.RepeatableString("host", h.Host)
	return nil
}

func (h *HostArgs) Validate() error {
	return nil
}

type RecordFilterArgs struct {
	HostArgs
	User              []string
	ExcludeUser       []string
	Command           []string
	ExcludeCommand    []string
	ExcludeSystemJobs bool
	Job               []uint32
	ExcludeJob        []uint32
}

func (r *RecordFilterArgs) Add(fs *flag.FlagSet) {
	r.HostArgs.Add(fs)
	fs.Var(NewRepeatableString(&r.User), "user",
		"Select records for this `user`, \"-\" for all (repeatable) [default: command dependent]")
	fs.Var(NewRepeatableString(&r.User), "u", "Short for -user `user`")
	fs.Var(NewRepeatableString(&r.ExcludeUser), "exclude-user",
		"Exclude records where the `user` equals this string (repeatable) [default: none]")
	fs.Var(NewRepeatableString(&r.Command), "command",
		"Select records where the `command` equals this string (repeatable) [default: all]")
	fs.Var(NewRepeatableString(&r.ExcludeCommand), "exclude-command",
		"Exclude records where the `command` equals this string (repeatable) [default: none]")
	fs.BoolVar(&r.ExcludeSystemJobs, "exclude-system-jobs", false,
		"Exclude records for system jobs (uid < 1000)")
	fs.Var(NewRepeatableUint32(&r.Job), "job",
		"Select records for this `job` ID (repeatable) [default: all]")
	fs.Var(NewRepeatableUint32(&r.Job), "j", "Short for -job `job`")
	fs.Var(NewRepeatableUint32(&r.ExcludeJob), "exclude-job",
		"Exclude jobs where the `job` ID equals this ID (repeatable) [default: none]")
}

func (r *RecordFilterArgs) ReifyForRemote(x *Reifier) error {
	e := r.HostArgs.ReifyForRemote(x)
	x.RepeatableString("user", r.User)
	x.RepeatableString("exclude-user", r.ExcludeUser)
	x.RepeatableString("command", r.Command)
	x.RepeatableString("exclude-command", r.ExcludeCommand)
	x.Bool("exclude-system-jobs", r.ExcludeSystemJobs)
	x.RepeatableUint32("job", r.Job)
	x.RepeatableUint32("exclude-job", r.ExcludeJob)
	return e
}

func (r *RecordFilterArgs) Validate() error {
	return r.HostArgs.Validate()
}

func (rfa *RecordFilterArgs) DefaultUserFilters() (allUsers, skipSystemUsers, determined bool) {
	if len(rfa.Job) > 0 {
		// `--job=...` implies `--user=-` b/c the job also implies a user.
		allUsers, skipSystemUsers = true, true
		determined = true
	} else if len(rfa.ExcludeUser) > 0 {
		// `--exclude-user=...` implies `--user=-` b/c the only sane way to include
		// many users so that some can be excluded is by also specifying `--users=-`.
		allUsers, skipSystemUsers = true, false
		determined = true
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Config file

type ConfigFileArgs struct {
	ConfigFilename string
}

func (cfa *ConfigFileArgs) Add(fs *flag.FlagSet) {
	fs.StringVar(&cfa.ConfigFilename, "config-file", "",
		"A `filename` for a file holding JSON data with system information, for when we\n"+
			"want to print or use system-relative values [default: none]")
}

func (cfa *ConfigFileArgs) ReifyForRemote(x *Reifier) error {
	if cfa.ConfigFilename != "" {
		return errors.New("-config-file can't be specified remotely")
	}
	return nil
}

func (cfa *ConfigFileArgs) Validate() error {
	if cfa.ConfigFilename != "" {
		cfa.ConfigFilename = path.Clean(cfa.ConfigFilename)
	}
	return nil
}

func (cfa *ConfigFileArgs) ConfigFile() string {
	return cfa.ConfigFilename
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Shared for all the analysis commands.  Some commands don't need the config file directly but
// it is required for caching data.

type SharedArgs struct {
	DevArgs
	SourceArgs
	RecordFilterArgs
	VerboseArgs
	ConfigFileArgs
}

func (sa *SharedArgs) SharedFlags() *SharedArgs {
	return sa
}

func (s *SharedArgs) Add(fs *flag.FlagSet) {
	s.DevArgs.Add(fs)
	s.SourceArgs.Add(fs)
	s.RecordFilterArgs.Add(fs)
	s.ConfigFileArgs.Add(fs)
	fs.BoolVar(&s.Verbose, "v", false, "Print verbose diagnostics to stderr")
	// The Rust version allows both -v and --verbose
	fs.BoolVar(&s.Verbose, "verbose", false, "Print verbose diagnostics to stderr")
}

func (s *SharedArgs) ReifyForRemote(x *Reifier) error {
	// We don't forward s.Verbose, it's mostly useful locally, and ideally sonalyzed should redact
	// it on the remote end to avoid revealing internal data (it does not, and indeed would require
	// the argument to be named "verbose" to work).
	return errors.Join(
		s.DevArgs.ReifyForRemote(x),
		s.SourceArgs.ReifyForRemote(x),
		s.RecordFilterArgs.ReifyForRemote(x),
		s.ConfigFileArgs.ReifyForRemote(x),
	)
}

func (s *SharedArgs) Validate() error {
	return errors.Join(
		s.DevArgs.Validate(),
		s.SourceArgs.Validate(),
		s.RecordFilterArgs.Validate(),
		s.VerboseArgs.Validate(),
		s.ConfigFileArgs.Validate(),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Repeatable arguments.  If we get one more of these types we should parameterize.

type RepeatableString struct {
	xs *[]string
}

func NewRepeatableString(xs *[]string) *RepeatableString {
	return &RepeatableString{xs}
}

func (rs *RepeatableString) String() string {
	if rs == nil || rs.xs == nil {
		return ""
	}
	return strings.Join(*rs.xs, ",")
}

func (rs *RepeatableString) Set(s string) error {
	// String arguments can't be comma-separated because host patterns such as 'ml[1,2],ml9' would
	// not really work without heroic effort.
	if *rs.xs == nil {
		*rs.xs = []string{s}
	} else {
		*rs.xs = append(*rs.xs, s)
	}
	return nil
}

type RepeatableUint32 struct {
	xs *[]uint32
}

func NewRepeatableUint32(xs *[]uint32) *RepeatableUint32 {
	return &RepeatableUint32{xs}
}

func (rs *RepeatableUint32) String() string {
	if rs == nil || rs.xs == nil {
		return ""
	}
	s := ""
	for _, v := range *rs.xs {
		if s != "" {
			s += ","
		}
		s += fmt.Sprint(v)
	}
	return s
}

func (rs *RepeatableUint32) Set(s string) error {
	// It's OK to allow comma-separated integer options since there is no parsing ambiguity.
	ys := strings.Split(s, ",")
	ws := make([]uint32, 0, len(ys))
	for _, y := range ys {
		if y == "" {
			return errors.New("Empty string")
		}
		n, err := strconv.ParseUint(y, 10, 32)
		if err != nil {
			return err
		}
		ws = append(ws, uint32(n))
	}
	if *rs.xs == nil {
		*rs.xs = ws
	} else {
		*rs.xs = append(*rs.xs, ws...)
	}
	return nil
}
