package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	. "sonalyze/common"
	. "sonalyze/table"
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

func (d *DevArgs) Add(fs *CLI) {
	if devArgs {
		fs.Group("development")
		fs.StringVar(&d.CpuProfile, "cpuprofile", "",
			"(Development) write cpu profile to `filename`")
	}
}

func (d *DevArgs) ReifyForRemote(x *ArgReifier) error {
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

func (va *VerboseArgs) Add(fs *CLI) {
	fs.Group("development")
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

func (dd *DataDirArgs) Add(fs *CLI) {
	fs.Group("local-data-source")
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

type RemotingArgsNoCluster struct {
	Remote   string
	AuthFile string

	Remoting bool
}

func (ra *RemotingArgsNoCluster) Add(fs *CLI) {
	fs.Group("remote-data-source")
	fs.StringVar(&ra.Remote, "remote", "",
		"Select a remote `url` to serve the query [default: none].")
	fs.StringVar(&ra.AuthFile, "auth-file", "",
		"Provide a `file` on username:password or netrc format [default: none].  For use with -remote.")
}

func (ra *RemotingArgsNoCluster) Validate() error {
	if ra.Remote != "" {
		ra.Remoting = true
	}
	return nil
}

func (va *RemotingArgsNoCluster) RemotingFlags() *RemotingArgsNoCluster {
	return va
}

type RemotingArgs struct {
	RemotingArgsNoCluster
	Cluster string
}

func (ra *RemotingArgs) Add(fs *CLI) {
	ra.RemotingArgsNoCluster.Add(fs)
	fs.Group("remote-data-source")
	fs.StringVar(&ra.Cluster, "cluster", "",
		"Select the cluster `name` for which we want data [default: none].  For use with -remote.")
}

func (ra *RemotingArgs) Validate() error {
	if ra.Remote != "" || ra.Cluster != "" {
		if ra.Remote == "" || ra.Cluster == "" {
			return errors.New("-remote and -cluster must be used together")
		}
		ra.Remoting = true
	}
	return nil
}

func (va *RemotingArgs) RemotingFlags() *RemotingArgsNoCluster {
	return &va.RemotingArgsNoCluster
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

	FromDateStr string
	ToDateStr   string
}

func (s *SourceArgs) Add(fs *CLI) {
	s.DataDirArgs.Add(fs)
	s.RemotingArgs.Add(fs)
	fs.Group("record-filter")
	fs.StringVar(&s.FromDateStr, "from", "",
		"Select records by this `time` and later.  Format can be YYYY-MM-DD, or Nd or Nw\n"+
			"signifying N days or weeks ago [default: 1d, ie 1 day ago]")
	fs.StringVar(&s.FromDateStr, "f", "", "Short for -from `time`")
	fs.StringVar(&s.ToDateStr, "to", "",
		"Select records by this `time` and earlier.  Format can be YYYY-MM-DD, or Nd or Nw\n"+
			"signifying N days or weeks ago [default: now]")
	fs.StringVar(&s.ToDateStr, "t", "", "Short for -to `time`")
}

func (s *SourceArgs) ReifyForRemote(x *ArgReifier) error {
	// RemotingArgs don't have ReifyForRemote

	// Validate() has already checked that DataDir, LogFiles, Remote, Cluster, and AuthFile are
	// consistent for remote or local execution; none of those except Cluster is forwarded.
	x.String("cluster", s.Cluster)
	x.String("from", s.FromDateStr)
	x.String("to", s.ToDateStr)
	return nil
}

func (s *SourceArgs) SetRestArguments(args []string) {
	s.LogFiles = args
}

func (s *SourceArgs) Validate() error {
	env_auth := os.Getenv("SONALYZE_AUTH") != ""
	switch {
	case len(s.LogFiles) > 0 || s.DataDir != "":
		// no action
	case s.Remote != "" || s.Cluster != "" || (env_auth || s.AuthFile != ""):
		ApplyDefault(&s.Remote, DataSourceRemote)
		if !env_auth {
			ApplyDefault(&s.AuthFile, DataSourceAuthFile)
		}
		ApplyDefault(&s.Cluster, DataSourceCluster)
	default:
		// There are no remoting args and no data dir args and no logfiles, so apply the ones we
		// have but error out if we have defaults for both.
		if (HasDefault(DataSourceRemote) ||
			(env_auth || HasDefault(DataSourceAuthFile)) ||
			HasDefault(DataSourceCluster)) &&
			HasDefault(DataSourceDataDir) {
			return errors.New("No data source, but defaults for both remoting and data directory")
		}
		if ApplyDefault(&s.DataDir, DataSourceDataDir) {
			// no action
		} else {
			ApplyDefault(&s.Remote, DataSourceRemote)
			if !env_auth {
				ApplyDefault(&s.AuthFile, DataSourceAuthFile)
			}
			ApplyDefault(&s.Cluster, DataSourceCluster)
		}
	}
	ApplyDefault(&s.FromDateStr, DataSourceFrom)
	ApplyDefault(&s.ToDateStr, DataSourceTo)

	err := s.RemotingArgs.Validate()
	if err != nil {
		return err
	}

	if s.Remoting {
		// If remoting then no local data sources are allowed, so don't compute default data dirs by
		// calling Validate(), it would confuse the matter - just disallow explicit values.  (This
		// is a small abstraction leak.)
		if s.DataDir != "" {
			return errors.New("-data-dir may not be used with -remote or -cluster")
		}
		if len(s.LogFiles) > 0 {
			return errors.New("-- logfile ... may not be used with -remote or -cluster")
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
			return errors.New("Required -data-dir or -- logfile ...")
		}
	}

	// The song and dance with `HaveFrom` and `HaveTo` is this: when a list of files is present then
	// `-from` and `-to` are inferred from the file contents, so long as they are not present on the
	// command line.

	now := time.Now().UTC()
	if s.FromDateStr != "" {
		var err error
		s.FromDate, err = ParseRelativeDateUtc(now, s.FromDateStr, false)
		if err != nil {
			return fmt.Errorf("Invalid -from argument %s", s.FromDateStr)
		}
		s.HaveFrom = true
	} else {
		s.FromDate = now.AddDate(0, 0, -1)
		s.HaveFrom = len(s.LogFiles) == 0
	}

	if s.ToDateStr != "" {
		var err error
		s.ToDate, err = ParseRelativeDateUtc(now, s.ToDateStr, true)
		if err != nil {
			return fmt.Errorf("Invalid -to argument %s", s.ToDateStr)
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

func (args *SourceArgs) InterpretFromToWithBounds(bounds Timebounds) (int64, int64) {
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

func (h *HostArgs) Add(fs *CLI) {
	fs.Group("record-filter")
	fs.Var(NewRepeatableStringNoCommas(&h.Host), "host",
		"Select records for this `host` (repeatable) [default: all]")
}

func (h *HostArgs) ReifyForRemote(x *ArgReifier) error {
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

func (r *RecordFilterArgs) Add(fs *CLI) {
	r.HostArgs.Add(fs)
	fs.Group("record-filter")
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

func (r *RecordFilterArgs) ReifyForRemote(x *ArgReifier) error {
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
// Query arguments

type QueryArgs struct {
	QueryStmt   string
	ParsedQuery PNode
}

func (qa *QueryArgs) Add(fs *CLI) {
	fs.Group("query")
	fs.StringVar(&qa.QueryStmt, "q", "", "A query expression")
}

func (qa *QueryArgs) ReifyForRemote(x *ArgReifier) error {
	x.String("q", qa.QueryStmt)
	return nil
}

func (qa *QueryArgs) Validate() (err error) {
	if qa.QueryStmt != "" {
		qa.ParsedQuery, err = ParseQuery(qa.QueryStmt)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Config file

type ConfigFileArgs struct {
	ConfigFilename string
}

func (cfa *ConfigFileArgs) Add(fs *CLI) {
	fs.Group("local-data-source")
	fs.StringVar(&cfa.ConfigFilename, "config-file", "",
		"A `filename` for a file holding JSON data with system information, for when we\n"+
			"want to print or use system-relative values [default: none]")
}

func (cfa *ConfigFileArgs) ReifyForRemote(x *ArgReifier) error {
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
// Format arguments - same logic for most consumers.

type FormatArgs struct {
	// Print args
	Fmt string

	// Synthesized and other
	PrintFields []FieldSpec
	PrintOpts   *FormatOptions
}

func (fa *FormatArgs) Add(fs *CLI) {
	fs.Group("printing")
	fs.StringVar(&fa.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (fa *FormatArgs) ReifyForRemote(x *ArgReifier) error {
	x.String("fmt", fa.Fmt)
	return nil
}

func ValidateFormatArgs[T any](
	fa *FormatArgs,
	defaultFields string,
	formatters map[string]Formatter[T],
	aliases map[string][]string,
	def DefaultFormat,
) error {
	var err error
	var others map[string]bool
	fa.PrintFields, others, err = ParseFormatSpec(defaultFields, fa.Fmt, formatters, aliases)
	if err == nil && len(fa.PrintFields) == 0 {
		err = errors.New("No valid output fields were selected in format string")
	}
	fa.PrintOpts = StandardFormatOptions(others, def)
	return err
}

func NeedsConfig[T any](formatters map[string]Formatter[T], fields []FieldSpec) bool {
	for _, f := range fields {
		if probe, found := formatters[f.Name]; found {
			if probe.NeedsConfig {
				return true
			}
		}
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Repeatable arguments.

// Some string arguments can't be comma-separated because host patterns such as 'ml[1,2],ml9' would
// not really work without heroic effort.

type RepeatableStringNoCommas struct {
	xs *[]string
}

func NewRepeatableStringNoCommas(xs *[]string) *RepeatableStringNoCommas {
	return &RepeatableStringNoCommas{xs}
}

func (rs *RepeatableStringNoCommas) String() string {
	if rs == nil || rs.xs == nil {
		return ""
	}
	return strings.Join(*rs.xs, ",")
}

func (rs *RepeatableStringNoCommas) Set(s string) error {
	if *rs.xs == nil {
		*rs.xs = []string{s}
	} else {
		*rs.xs = append(*rs.xs, s)
	}
	return nil
}

type RepeatableCommaSeparated[T any] struct {
	xs         *[]T
	fromString func(string) (T, error)
}

func (rs *RepeatableCommaSeparated[T]) String() string {
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

func (rs *RepeatableCommaSeparated[T]) Set(s string) error {
	ys := strings.Split(s, ",") // OK: "" is ruled out below
	ws := make([]T, 0, len(ys))
	for _, y := range ys {
		if y == "" {
			return errors.New("Empty string is an invalid argument")
		}
		n, err := rs.fromString(y)
		if err != nil {
			return err
		}
		ws = append(ws, n)
	}
	if *rs.xs == nil {
		*rs.xs = ws
	} else {
		*rs.xs = append(*rs.xs, ws...)
	}
	return nil
}

type RepeatableUint32 = RepeatableCommaSeparated[uint32]

func NewRepeatableUint32(xs *[]uint32) *RepeatableUint32 {
	return &RepeatableCommaSeparated[uint32]{
		xs,
		func(s string) (uint32, error) {
			n, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return uint32(n), nil
		},
	}
}

type RepeatableString = RepeatableCommaSeparated[string]

func NewRepeatableString(xs *[]string) *RepeatableString {
	return &RepeatableString{
		xs,
		func(s string) (string, error) {
			return s, nil
		},
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Shared args for analysis commands that work on sonar samples.  (Some commands don't need the config
// file for data processing but it is required for caching data.)

type SampleAnalysisArgs struct {
	DevArgs
	SourceArgs
	QueryArgs
	RecordFilterArgs
	ConfigFileArgs
	VerboseArgs
}

func (sa *SampleAnalysisArgs) SampleAnalysisFlags() *SampleAnalysisArgs {
	return sa
}

func (s *SampleAnalysisArgs) Add(fs *CLI) {
	s.DevArgs.Add(fs)
	s.SourceArgs.Add(fs)
	s.QueryArgs.Add(fs)
	s.RecordFilterArgs.Add(fs)
	s.ConfigFileArgs.Add(fs)
	s.VerboseArgs.Add(fs)
}

func (s *SampleAnalysisArgs) ReifyForRemote(x *ArgReifier) error {
	// We don't forward s.Verbose, it's mostly useful locally, and ideally sonalyzed should redact
	// it on the remote end to avoid revealing internal data.
	return errors.Join(
		s.DevArgs.ReifyForRemote(x),
		s.SourceArgs.ReifyForRemote(x),
		s.QueryArgs.ReifyForRemote(x),
		s.RecordFilterArgs.ReifyForRemote(x),
		s.ConfigFileArgs.ReifyForRemote(x),
	)
}

func (s *SampleAnalysisArgs) Validate() error {
	return errors.Join(
		s.DevArgs.Validate(),
		s.SourceArgs.Validate(),
		s.QueryArgs.Validate(),
		s.RecordFilterArgs.Validate(),
		s.ConfigFileArgs.Validate(),
		s.VerboseArgs.Validate(),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Shared args for analysis commands that work on sonar per-host data.
//
// Almost SampleAnalysisArgs, but HostArgs instead of RecordFilterArgs

type HostAnalysisArgs struct {
	DevArgs
	SourceArgs
	QueryArgs
	HostArgs
	ConfigFileArgs
	VerboseArgs
}

func (sa *HostAnalysisArgs) HostAnalysisFlags() *HostAnalysisArgs {
	return sa
}

func (s *HostAnalysisArgs) Add(fs *CLI) {
	s.DevArgs.Add(fs)
	s.SourceArgs.Add(fs)
	s.QueryArgs.Add(fs)
	s.HostArgs.Add(fs)
	s.ConfigFileArgs.Add(fs)
	s.VerboseArgs.Add(fs)
}

func (s *HostAnalysisArgs) ReifyForRemote(x *ArgReifier) error {
	// We don't forward s.Verbose, it's mostly useful locally, and ideally sonalyzed should redact
	// it on the remote end to avoid revealing internal data.
	return errors.Join(
		s.DevArgs.ReifyForRemote(x),
		s.SourceArgs.ReifyForRemote(x),
		s.QueryArgs.ReifyForRemote(x),
		s.HostArgs.ReifyForRemote(x),
		s.ConfigFileArgs.ReifyForRemote(x),
	)
}

func (s *HostAnalysisArgs) Validate() error {
	return errors.Join(
		s.DevArgs.Validate(),
		s.SourceArgs.Validate(),
		s.QueryArgs.Validate(),
		s.HostArgs.Validate(),
		s.ConfigFileArgs.Validate(),
		s.VerboseArgs.Validate(),
	)
}
