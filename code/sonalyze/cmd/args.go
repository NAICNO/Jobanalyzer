package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"go-utils/options"
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
// The data source.  This is fairly elaborate to cover multiple use cases.
//
// The data source can be "remote" or local.  If it is remote, then the command is forwarded to the
// remote server and executed on a local data store there, ie, a "remote" source is just a REST
// call.
//
// To specify a remote source, use -remote.  This optionally takes -auth-file and by default (for
// historical reasons) requires -cluster, though commands can use DBArgNoCluster to specify that
// -cluster should not be accepted.  Also for historical reasons, -cluster implies -remote.
//
// If the source is remote, then the env var SONALYZE_AUTH can specify the value for -auth-file,
// this overrides anything specified elsewhere.
//
// If the source is remote, then values for the remote host and the auth file are fetched from
// ~/.sonalyze if they are not specified on the command line.
//
// If the source is not remote then it is local.
//
// If a -jobanalyzer-dir argument is present then it specifies the root location of the database and
// *all* data are found within that directory at known locations.  This precludes -data-dir,
// -config-file, -report-dir, and a file list.
//
// Otherwise, if -data-dir is present then that is a directory for a single cluster's sample,
// sysinfo, and slurm data store and the -config-file argument may provide cluster configuration
// data.  This precludes -report-dir and a file list.
//
// Otherwise, if -report-dir is present then that is a directory for a single cluster's
// generated-reports data store.  This precludes -data-dir and a file list.
//
// Otherwise, there should be a file list representing one kind of data for a single cluster, and
// the -config-file argument may provide cluster configuration data for that cluster.
//
// (In the past, there were environment-variable options for the data directory; these have been
// retired.)

type DatabaseArgs struct {
	Remote   string
	AuthFile string
	Remoting bool
	Cluster  string

	JobanalyzerDir string
	DataDir        string
	ReportDir      string
	ConfigFilename string
	LogFiles       []string

	options DBArgOptions
}

func (db *DatabaseArgs) ConfigFile() string {
	return db.ConfigFilename
}

type DBArgOptions struct {
	// Do not accept -cluster, as the command is not cluster-specific.
	OmitCluster bool

	// Include -report-dir, and require or compute (as needed) -report-dir for local execution, this
	// precludes -data-dir and file lists.
	IncludeReportDir bool

	// There is no database, do not open any data source, but handle remote execution.
	NoDatabase bool
}

func (db *DatabaseArgs) Add(fs *CLI, opts DBArgOptions) {
	db.options = opts

	fs.Group("remote-data-source")
	fs.StringVar(&db.Remote, "remote", "",
		"Select a remote `url` to serve the query [default: none].")
	fs.StringVar(&db.AuthFile, "auth-file", "",
		"Provide a `file` on username:password or netrc format [default: none].  For use with -remote.")
	if opts.OmitCluster {
		fs.StringVar(&db.Cluster, "cluster", "",
			"Select the cluster `name` for which we want data [default: none].  For use with -remote.")
	}

	fs.Group("local-data-source")
	fs.StringVar(&db.JobanalyzerDir, "jobanalyzer-dir", "",
		"Jobanalyzer root `directory`, precludes all other local data source arguments.")
	fs.StringVar(&db.DataDir, "data-dir", "",
		"Select the root `directory` for log files [default: none]")
	fs.StringVar(&db.DataDir, "data-path", "", "Alias for -data-dir `directory`")
	fs.StringVar(&db.ConfigFilename, "config-file", "",
		"A `filename` for a file holding JSON data with system information, for when we\n"+
			"want to print or use system-relative values [default: none]")
	if opts.IncludeReportDir {
		fs.StringVar(
			&db.ReportDir, "report-dir", "", "`directory-name` containing reports [default: none]")
	}
}

func (db *DatabaseArgs) RemotingFlags() RemotingFlags {
	return RemotingFlags{
		Remoting: db.Remoting,
		AuthFile: db.AuthFile,
		Remote:   db.Remote,
	}
}

func (db *DatabaseArgs) SetRestArguments(args []string) {
	db.LogFiles = args
}

func (db *DatabaseArgs) Validate() error {
	// The trigger for remote execution is -remote or -cluster
	if db.Remote != "" || db.Cluster != "" {
		db.Remoting = true
	}

	// Basic validation of mutually exclusive situations
	switch {
	case db.Remoting:
		if db.JobanalyzerDir != "" {
			return errors.New("Remote execution precludes -jobanalyzer-dir")
		}
		if db.DataDir != "" {
			return errors.New("Remote execution precludes a -data-dir")
		}
		if db.ReportDir != "" {
			return errors.New("Remote execution precludes a -report-dir")
		}
		if db.ConfigFilename != "" {
			return errors.New("Remote execution precludes a -config-file")
		}
		if len(db.LogFiles) > 0 {
			return errors.New("Remote execution precludes a file list")
		}
	case db.JobanalyzerDir != "":
		if db.DataDir != "" {
			return errors.New("A -jobanalyzer-dir precludes a -data-dir")
		}
		if db.ReportDir != "" {
			return errors.New("A -jobanalyzer-dir precludes a -report-dir")
		}
		if db.ConfigFilename != "" {
			return errors.New("A -jobanalyzer-dir precludes a -config-file")
		}
		if len(db.LogFiles) > 0 {
			return errors.New("A -data-dir precludes a file list")
		}
	case db.DataDir != "":
		if db.ReportDir != "" {
			return errors.New("A -data-dir a -report-dir")
		}
		if len(db.LogFiles) > 0 {
			return errors.New("A -data-dir precludes a file list")
		}
	case db.ReportDir != "":
		if len(db.LogFiles) > 0 {
			return errors.New("A -report-dir precludes a file list")
		}
	}

	// TODO: Not sure what to do about -report-dir yet: do we need to compute something?  Or is it
	// hidden inside the database eventually?

	// TODO: A bare -config-file is also a data source - when you're running the `config` command.
	// But that's a pretty special case.  It's like -report-dir: it's only meaningful when running
	// `report`.  Possibly this next test needs a few more conditions.

	if !db.options.NoDatabase {
		if !db.Remoting &&
			db.JobanalyzerDir == "" &&
			db.DataDir == "" &&
			db.ReportDir == "" &&
			db.ConfigFilename == "" &&
			len(db.LogFiles) == 0 {
			return errors.New("No data source provided")
		}
	}

	// Apply defaults for the remote data source
	if db.Remoting {
		ApplyDefault(&db.Remote, DataSourceRemote)
		if os.Getenv("SONALYZE_AUTH") == "" {
			ApplyDefault(&db.AuthFile, DataSourceAuthFile)
		}
		ApplyDefault(&db.Cluster, DataSourceCluster)
	}

	// Clean all local names and check that they exist, for better error reporting.
	var e1, e2, e3, e4, e5, e6 error
	if db.JobanalyzerDir != "" {
		db.JobanalyzerDir, e1 = options.RequireDirectory(db.JobanalyzerDir, "-jobanalyzer-dir")
	}

	if db.DataDir != "" {
		db.DataDir, e2 = options.RequireDirectory(db.DataDir, "-data-dir")
	}

	if db.ReportDir != "" {
		db.ReportDir, e3 = options.RequireDirectory(db.ReportDir, "-report-dir")
	}

	if db.ConfigFilename != "" {
		db.ConfigFilename, e4 = options.RequireFile(db.ConfigFilename, "-config-file")
	}

	for i := range db.LogFiles {
		var e error
		db.LogFiles[i], e = options.RequireFile(db.LogFiles[i], "")
		if e != nil {
			e5 = errors.New("No such input file: " + db.LogFiles[i])
			break
		}
	}

	if db.AuthFile != "" {
		db.AuthFile, e6 = options.RequireFile(db.AuthFile, "-auth-file")
	}

	return errors.Join(e1, e2, e3, e4, e5, e6)
}

func (db *DatabaseArgs) ReifyForRemote(x *ArgReifier) error {
	// Validate() has already checked that JobanalyzerDir, DataDir, ReportDir, LogFiles, Remote,
	// Cluster, and AuthFile are consistent for remote and local execution.  None of those except
	// Cluster is forwarded for remote execution.
	x.String("cluster", db.Cluster)
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// SourceArgs pertain to source file location and initial filtering-by-location, though the
// -from/-to arguments are also used to filter records.

type SourceArgs struct {
	DatabaseArgs
	HaveFrom bool
	FromDate time.Time
	HaveTo   bool
	ToDate   time.Time

	FromDateStr string
	ToDateStr   string
}

func (s *SourceArgs) Add(fs *CLI) {
	s.DatabaseArgs.Add(fs, DBArgOptions{})
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
	x.String("from", s.FromDateStr)
	x.String("to", s.ToDateStr)
	return s.DatabaseArgs.ReifyForRemote(x)
}

func (s *SourceArgs) Validate() error {
	// The song and dance with `HaveFrom` and `HaveTo` is this: when a list of files is present then
	// `-from` and `-to` are inferred from the file contents, so long as they are not present on the
	// command line.

	ApplyDefault(&s.FromDateStr, DataSourceFrom)
	ApplyDefault(&s.ToDateStr, DataSourceTo)

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

	return s.DatabaseArgs.Validate()
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
	)
}

func (s *SampleAnalysisArgs) Validate() error {
	return errors.Join(
		s.DevArgs.Validate(),
		s.SourceArgs.Validate(),
		s.QueryArgs.Validate(),
		s.RecordFilterArgs.Validate(),
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
	)
}

func (s *HostAnalysisArgs) Validate() error {
	return errors.Join(
		s.DevArgs.Validate(),
		s.SourceArgs.Validate(),
		s.QueryArgs.Validate(),
		s.HostArgs.Validate(),
		s.VerboseArgs.Validate(),
	)
}
