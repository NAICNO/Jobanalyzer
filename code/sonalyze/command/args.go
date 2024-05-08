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

	"go-utils/minmax"
	"sonalyze/common"
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
// SourceArgs pertain to source file location and initial filtering-by-location, though the
// -from/-to arguments are also used to filter records.

type SourceArgs struct {
	DataDir  string
	HaveFrom bool
	FromDate time.Time
	HaveTo   bool
	ToDate   time.Time
	Remote   string
	Cluster  string
	AuthFile string
	LogFiles []string

	Remoting    bool
	fromDateStr string
	toDateStr   string
}

func (s *SourceArgs) Add(fs *flag.FlagSet) {
	fs.StringVar(&s.DataDir, "data-dir", "",
		"Select the root `directory` for log files [default: $SONAR_ROOT or $HOME/sonar/data]")
	fs.StringVar(&s.DataDir, "data-path", "", "Alias for -data-dir `directory`")
	fs.StringVar(&s.fromDateStr, "from", "",
		"Select records by this `time` and later.  Format can be YYYY-MM-DD, or Nd or Nw\n"+
			"signifying N days or weeks ago [default: 1d, ie 1 day ago]")
	fs.StringVar(&s.fromDateStr, "f", "", "Short for -from `time`")
	fs.StringVar(&s.toDateStr, "to", "",
		"Select records by this `time` and earlier.  Format can be YYYY-MM-DD, or Nd or Nw\n"+
			"signifying N days or weeks ago [default: now]")
	fs.StringVar(&s.toDateStr, "t", "", "Short for -to `time`")
	fs.StringVar(&s.Remote, "remote", "",
		"Select a remote `url` to serve the query [default: none].  Requires -cluster.")
	fs.StringVar(&s.Cluster, "cluster", "",
		"Select the cluster `name` for which we want data [default: none].  For use with -remote.")
	fs.StringVar(&s.AuthFile, "auth-file", "",
		"Provide a `file` with username:password [default: none].  For use with -remote.")
}

func (s *SourceArgs) ReifyForRemote(x *Reifier) error {
	// Validate() has already checked that DataDir, LogFiles, Remote, Cluster, and AuthFile are
	// consistent for remote or local execution; none of those except Cluster is forwarded.
	x.String("cluster", s.Cluster)
	x.String("from", s.fromDateStr)
	x.String("to", s.toDateStr)
	return nil
}

func (s *SourceArgs) Validate() error {
	// Compute and clean the dataDir and clean any logfiles.  If we have neither logfiles nor
	// dataDir then signal an error.

	s.Remoting = s.Remote != "" || s.Cluster != ""

	if !s.Remoting {
		if s.DataDir != "" {
			s.DataDir = path.Clean(s.DataDir)
		} else if d := os.Getenv("SONAR_ROOT"); d != "" {
			s.DataDir = path.Clean(d)
		} else if d := os.Getenv("HOME"); d != "" {
			s.DataDir = path.Clean(path.Join(d, "/sonar/data"))
		}
	}

	if len(s.LogFiles) > 0 {
		for i := 0; i < len(s.LogFiles); i++ {
			s.LogFiles[i] = path.Clean(s.LogFiles[i])
		}
	} else if s.DataDir == "" && !s.Remoting {
		return fmt.Errorf("Required -data-dir or -- logfile ...")
	}

	if s.Remoting {
		if s.DataDir != "" {
			return fmt.Errorf("-data-dir may not be used with -remote or -cluster")
		}
		if len(s.LogFiles) > 0 {
			return fmt.Errorf("-- logfile ... may not be used with -remote or -cluster")
		}
		if s.Remote == "" || s.Cluster == "" {
			return fmt.Errorf("-remote and -cluster must be used together")
		}
	}

	// The song and dance with `HaveFrom` and `HaveTo` is this: when a list of files is present then
	// `-from` and `-to` are inferred from the file contents, so long as they are not present on the
	// command line.
	//
	// NOTE this uses our own version of ParseRelativeDate(), not the one in go-utils/time, for
	// compatibility with the Rust code.

	now := time.Now()
	if s.fromDateStr != "" {
		var err error
		s.FromDate, err = common.ParseRelativeDate(now, s.fromDateStr, false)
		if err != nil {
			return fmt.Errorf("Invalid -from argument %s", s.fromDateStr)
		}
		s.HaveFrom = true
	} else {
		s.FromDate = now.UTC().AddDate(0, 0, -1)
		s.HaveFrom = len(s.LogFiles) == 0
	}

	if s.toDateStr != "" {
		var err error
		s.ToDate, err = common.ParseRelativeDate(now, s.toDateStr, true)
		if err != nil {
			return fmt.Errorf("Invalid -to argument %s", s.toDateStr)
		}
		s.HaveTo = true
	} else {
		s.ToDate = now.UTC()
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
			from = minmax.MinInt64(from, v.Earliest)
		}
	}
	if args.HaveTo || len(bounds) == 0 {
		to = args.ToDate.Unix()
	} else {
		for _, v := range bounds {
			to = minmax.MaxInt64(to, v.Latest)
		}
	}
	return from, to
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// RecordFilterArgs pertain to including and excluding records by purely record-local criteria.  In
// addition to these, the -from / -to arguments of the SourceArgs are applied as record filters.

type RecordFilterArgs struct {
	Host              []string
	User              []string
	ExcludeUser       []string
	Command           []string
	ExcludeCommand    []string
	ExcludeSystemJobs bool
	Job               []uint32
	ExcludeJob        []uint32
}

func (r *RecordFilterArgs) Add(fs *flag.FlagSet) {
	fs.Var(newRepeatableString(&r.Host), "host", "Select records for this `host` (repeatable) [default: all]")
	fs.Var(newRepeatableString(&r.User), "user",
		"Select records for this `user`, \"-\" for all (repeatable) [default: command dependent]")
	fs.Var(newRepeatableString(&r.User), "u", "Short for -user `user`")
	fs.Var(newRepeatableString(&r.ExcludeUser), "exclude-user",
		"Exclude records where the `user` equals this string (repeatable) [default: none]")
	fs.Var(newRepeatableString(&r.Command), "command",
		"Select records where the `command` equals this string (repeatable) [default: all]")
	fs.Var(newRepeatableString(&r.ExcludeCommand), "exclude-command",
		"Exclude records where the `command` equals this string (repeatable) [default: none]")
	fs.BoolVar(&r.ExcludeSystemJobs, "exclude-system-jobs", false,
		"Exclude records for system jobs (uid < 1000)")
	fs.Var(newRepeatableUint32(&r.Job), "job",
		"Select records for this `job` ID (repeatable) [default: all]")
	fs.Var(newRepeatableUint32(&r.Job), "j", "Short for -job `job`")
	fs.Var(newRepeatableUint32(&r.ExcludeJob), "exclude-job",
		"Exclude jobs where the `job` ID equals this ID (repeatable) [default: none]")
}

func (r *RecordFilterArgs) ReifyForRemote(x *Reifier) error {
	x.RepeatableString("host", r.Host)
	x.RepeatableString("user", r.User)
	x.RepeatableString("exclude-user", r.ExcludeUser)
	x.RepeatableString("command", r.Command)
	x.RepeatableString("exclude-command", r.ExcludeCommand)
	x.Bool("exclude-system-jobs", r.ExcludeSystemJobs)
	x.RepeatableUint32("job", r.Job)
	x.RepeatableUint32("exclude-job", r.ExcludeJob)
	return nil
}

func (r *RecordFilterArgs) Validate() error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Shared for everyone

type SharedArgs struct {
	DevArgs
	SourceArgs
	RecordFilterArgs
	Verbose bool
}

func (sa *SharedArgs) Args() *SharedArgs {
	return sa
}

func (s *SharedArgs) Add(fs *flag.FlagSet) {
	s.DevArgs.Add(fs)
	s.SourceArgs.Add(fs)
	s.RecordFilterArgs.Add(fs)
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
	)
}

func (s *SharedArgs) Validate() error {
	return errors.Join(
		s.DevArgs.Validate(),
		s.SourceArgs.Validate(),
		s.RecordFilterArgs.Validate(),
	)
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
// Repeatable arguments.  If we get one more of these types we should parameterize.

type repeatableString struct {
	xs *[]string
}

func newRepeatableString(xs *[]string) *repeatableString {
	return &repeatableString{xs}
}

func (rs *repeatableString) String() string {
	if rs == nil || rs.xs == nil {
		return ""
	}
	return strings.Join(*rs.xs, ",")
}

func (rs *repeatableString) Set(s string) error {
	ys := strings.Split(s, ",")
	for _, y := range ys {
		if y == "" {
			return errors.New("Empty string")
		}
	}
	if *rs.xs == nil {
		*rs.xs = ys
	} else {
		*rs.xs = append(*rs.xs, ys...)
	}
	return nil
}

type repeatableUint32 struct {
	xs *[]uint32
}

func newRepeatableUint32(xs *[]uint32) *repeatableUint32 {
	return &repeatableUint32{xs}
}

func (rs *repeatableUint32) String() string {
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

func (rs *repeatableUint32) Set(s string) error {
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
