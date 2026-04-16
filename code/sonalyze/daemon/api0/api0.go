// The v0 REST API is a disciplined reimplementation of the query parts of the original sonalyze
// REST API.  In v0, the successful result of a GET is always a single JSON string that must be
// parsed by the consumer (for casual use just pipe it through `jq -r`), and when parsed should
// yield exactly the same text as the original API did.
//
// The v0 API is probably the right API for traditional sonalyze remoting, but for scripting we want
// to phase out this API in favor of the v1 API which returns proper JSON.
//
// There is no v0 API insertion point, as the original sonalyze insertion points are no longer
// supported - they handled only old data types, and not even all of those.  Use the v1 insertion
// points to insert data.

package api0

import (
	"context"
	"path"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"go-utils/auth"
	"sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/daemon/apiutil"
)

var (
	jobanalyzerDir   string
	databaseURI      string
	cmdlineHandler   cmd.CommandLineHandler
	getAuthenticator *auth.Authenticator
)

func SetupAPI(
	api huma.API,
	jobanalyzerDir_ string,
	databaseURI_ string,
	cmdlineHandler_ cmd.CommandLineHandler,
	getAuthenticator_ *auth.Authenticator,
) {
	jobanalyzerDir = jobanalyzerDir_
	databaseURI = databaseURI_
	cmdlineHandler = cmdlineHandler_
	getAuthenticator = getAuthenticator_
	grp := huma.NewGroup(api, "/api/v0")
	// WHEN UPDATING THESE, ALSO UPDATE SWITCH IN ../../application/command.go and HELP TEXT IN THE
	// SAME PLACE.
	addCard(grp)
	addCluster(grp)
	addConfig(grp)
	addDiskprof(grp)
	addGpu(grp)
	addJobs(grp)
	addLoad(grp)
	addMetadata(grp)
	addNode(grp)
	addNodeprof(grp)
	addProfile(grp)
	addSacct(grp)
	addSample(grp)
	addSnode(grp)
	addSpart(grp)
	addUptime(grp)
	addVersion(grp)
	// Omitting `add` because it was already obsolete; replaced by /api/v1/insert
	// Omitting `parse` because that's the old name for `sample`
	// Omitting `report` because it's obsolete, it was for dashboard-1
	// Omitting `top` for now because it is very limited
}

// Query commands.

type QueryResponse struct {
	// Body is a JSON string holding the output of the sonalyze command and it must be parsed, see
	// top comments.
	Body string
}

func addCard(api huma.API) {
	huma.Get(api, "/card", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"card",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams),
		)
	})
}

func addCluster(api huma.API) {
	huma.Get(api, "/cluster", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			QueryParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"cluster",
			input.Auth,
			collectAll(&input.QueryParams, &input.FormatParams),
		)
	})
}

func addConfig(api huma.API) {
	huma.Get(api, "/config", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"config",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams),
		)
	})
}

func addDiskprof(api huma.API) {
	huma.Get(api, "/diskprof", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"diskprof",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams),
		)
	})
}

func addGpu(api huma.API) {
	huma.Get(api, "/gpu", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
			GpuIndexParam // just in-line this if it only has this one use
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"gpu",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams, &input.GpuIndexParam),
		)
	})
}

func addJobs(api huma.API) {
	huma.Get(api, "/jobs", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			SampleAnalysisParams
			NoGpu          string `query:"no-gpu"`
			SomeGpu        string `query:"some-gpu"`
			Completed      string `query:"completed"`
			Running        string `query:"running"`
			Zombie         string `query:"zombie"`
			Partition      string `query:"partition"`
			Account        string `query:"account"`
			Reservation    string `query:"reservation"`
			State          string `query:"state"`
			GpuType        string `query:"gpu-type"`
			MinRuntimeSec  string `query:"min-runtime"`
			MergeAll       string `query:"merge-all"`
			MergeNone      string `query:"merge-none"`
			SacctFromSonar string `query:"sacct-from-sonar"`
			NumJobs        string `query:"numjobs"` // [sic!]
			MinSamples     string `query:"min-samples"`
			MinCpuAvg      string `query:"min-cpu-avg"`
			MinCpuPeak     string `query:"min-cpu-peak"`
			MaxCpuAvg      string `query:"max-cpu-avg"`
			MaxCpuPeak     string `query:"max-cpu-peak"`
			MinRcpuAvg     string `query:"min-rcpu-avg"`
			MinRcpuPeak    string `query:"min-rcpu-peak"`
			MaxRcpuAvg     string `query:"max-rcpu-avg"`
			MaxRcpuPeak    string `query:"max-rcpu-peak"`
			MinMemAvg      string `query:"min-mem-avg"`
			MinMemPeak     string `query:"min-mem-peak"`
			MinRmemAvg     string `query:"min-rmem-avg"`
			MinRmemPeak    string `query:"min-rmem-peak"`
			MinResAvg      string `query:"min-res-avg"`
			MinResPeak     string `query:"min-res-peak"`
			MinRresAvg     string `query:"min-rres-avg"`
			MinRresPeak    string `query:"min-rres-peak"`
			MinGpuAvg      string `query:"min-gpu-avg"`
			MinGpuPeak     string `query:"min-gpu-peak"`
			MaxGpuAvg      string `query:"max-gpu-avg"`
			MaxGpuPeak     string `query:"max-gpu-peak"`
			MinRgpuAvg     string `query:"min-rgpu-avg"`
			MinRgpuPeak    string `query:"min-rgpu-peak"`
			MaxRgpuAvg     string `query:"max-rgpu-avg"`
			MaxRgpuPeak    string `query:"max-rgpu-peak"`
			MinGpumemAvg   string `query:"min-gpumem-avg"`
			MinGpumemPeak  string `query:"min-gpumem-peak"`
			MinRgpumemAvg  string `query:"min-rgpumem-avg"`
			MinRgpumemPeak string `query:"min-rgpumem-peak"`
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"jobs",
			input.Auth,
			append(
				collectAll(&input.SampleAnalysisParams, &input.FormatParams),
				collect(
					"no-gpu", input.NoGpu,
					"some-gpu", input.SomeGpu,
					"completed", input.Completed,
					"running", input.Running,
					"zombie", input.Zombie,
					"partition", input.Partition,
					"account", input.Account,
					"reservation", input.Reservation,
					"state", input.State,
					"gpu-type", input.GpuType,
					"min-runtime", input.MinRuntimeSec,
					"merge-all", input.MergeAll,
					"merge-none", input.MergeNone,
					"sacct-from-sonar", input.SacctFromSonar,
					"numjobs", input.NumJobs,
					"min-samples", input.MinSamples,
					"min-cpu-avg", input.MinCpuAvg,
					"min-cpu-peak", input.MinCpuPeak,
					"max-cpu-avg", input.MaxCpuAvg,
					"max-cpu-peak", input.MaxCpuPeak,
					"min-rcpu-avg", input.MinRcpuAvg,
					"min-rcpu-peak", input.MinRcpuPeak,
					"max-rcpu-avg", input.MaxRcpuAvg,
					"max-rcpu-peak", input.MaxRcpuPeak,
					"min-mem-avg", input.MinMemAvg,
					"min-mem-peak", input.MinMemPeak,
					"min-rmem-avg", input.MinRmemAvg,
					"min-rmem-peak", input.MinRmemPeak,
					"min-res-avg", input.MinResAvg,
					"min-res-peak", input.MinResPeak,
					"min-rres-avg", input.MinRresAvg,
					"min-rres-peak", input.MinRresPeak,
					"min-gpu-avg", input.MinGpuAvg,
					"min-gpu-peak", input.MinGpuPeak,
					"max-gpu-avg", input.MaxGpuAvg,
					"max-gpu-peak", input.MaxGpuPeak,
					"min-rgpu-avg", input.MinRgpuAvg,
					"min-rgpu-peak", input.MinRgpuPeak,
					"max-rgpu-avg", input.MaxRgpuAvg,
					"max-rgpu-peak", input.MaxRgpuPeak,
					"min-gpumem-avg", input.MinGpumemAvg,
					"min-gpumem-peak", input.MinGpumemPeak,
					"min-rgpumem-avg", input.MinRgpumemAvg,
					"min-rgpumem-peak", input.MinRgpumemPeak,
				)...,
			),
		)
	})
}

func addLoad(api huma.API) {
	huma.Get(api, "/load", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			SampleAnalysisParams
			FormatParams
			Hourly     string `query:"hourly"`
			HalfHourly string `query:"half-hourly"`
			Daily      string `query:"daily"`
			HalfDaily  string `query:"half-daily"`
			Weekly     string `query:"weekly"`
			None       string `query:"none"`
			Group      string `query:"group"`
			All        string `query:"all"`
			Last       string `query:"last"`
			Compact    string `query:"compact"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"load",
			input.Auth,
			append(
				collectAll(&input.SampleAnalysisParams, &input.FormatParams),
				collect(
					"hourly", input.Hourly,
					"half-hourly", input.HalfHourly,
					"daily", input.Daily,
					"half-daily", input.HalfDaily,
					"weekly", input.Weekly,
					"none", input.None,
					"group", input.Group,
					"all", input.All,
					"last", input.Last,
					"compact", input.Compact,
				)...,
			),
		)
	})
}

func addMetadata(api huma.API) {
	huma.Get(api, "/metadata", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			SampleAnalysisParams
			FormatParams
			MergeByHostAndJob string `query:"merge-by-host-and-job"`
			MergeByJob        string `query:"merge-by-job"`
			Times             string `query:"times"`
			Files             string `query:"files"`
			Bounds            string `query:"bounds"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"metadata",
			input.Auth,
			append(
				collectAll(&input.SampleAnalysisParams, &input.FormatParams),
				collect(
					"merge-by-host-and-job", input.MergeByHostAndJob,
					"merge-by-job", input.MergeByJob,
					"times", input.Times,
					"files", input.Files,
					"bounds", input.Bounds,
				)...,
			),
		)
	})
}

func addNode(api huma.API) {
	huma.Get(api, "/node", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
			Newest string `query:"newest"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"node",
			input.Auth,
			append(
				collectAll(&input.HostAnalysisParams, &input.FormatParams),
				collect("newest", input.Newest)...,
			),
		)
	})
}

func addNodeprof(api huma.API) {
	huma.Get(api, "/nodeprof", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"nodeprof",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams),
		)
	})
}

func addProfile(api huma.API) {
	huma.Get(api, "/profile", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			SampleAnalysisParams
			FormatParams
			Max    string `query:"max"`
			Bucket string `query:"bucket"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"parse",
			input.Auth,
			append(
				collectAll(&input.SampleAnalysisParams, &input.FormatParams),
				collect("max", input.Max, "bucket", input.Bucket)...,
			),
		)
	})
}

func addSacct(api huma.API) {
	huma.Get(api, "/sacct", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
			States           string `query:"state"`
			Users            string `query:"user"`
			Accounts         string `query:"account"`
			Partitions       string `query:"partition"`
			Jobs             string `query:"job"`
			All              string `query:"all"`
			MinRuntime       string `query:"min-runtime"`
			MaxRuntime       string `query:"max-runtime"`
			MinReservedMem   string `query:"min-reserved-mem"`
			MaxReservedMem   string `query:"max-reserved-mem"`
			MinReservedCores string `query:"min-reserved-cores"`
			MaxReservedCores string `query:"max-reserved-cores"`
			SomeGPU          string `query:"some-gpu"`
			NoGPU            string `query:"no-gpu"`
			Regular          string `query:"regular"`
			Array            string `query:"array"`
			Het              string `query:"het"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"sacct",
			input.Auth,
			append(
				collectAll(&input.HostAnalysisParams, &input.FormatParams),
				collect(
					"state", input.States,
					"user", input.Users,
					"account", input.Accounts,
					"partition", input.Partitions,
					"job", input.Jobs,
					"all", input.All,
					"min-runtime", input.MinRuntime,
					"max-runtime", input.MaxRuntime,
					"min-reserved-mem", input.MinReservedMem,
					"max-reserved-mem", input.MaxReservedMem,
					"min-reserved-cores", input.MinReservedCores,
					"max-reserved-cores", input.MaxReservedCores,
					"some-gpu", input.SomeGPU,
					"no-gpu", input.NoGPU,
					"regular", input.Regular,
					"array", input.Array,
					"het", input.Het,
				)...,
			),
		)
	})
}

func addSample(api huma.API) {
	huma.Get(api, "/parse", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			SampleAnalysisParams
			FormatParams
			MergeByHostAndJob string `query:"merge-by-host-and-job"`
			MergeByJob        string `query:"merge-by-job"`
			Clean             string `query:"clean"`
			LastN             string `query:"last"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"parse",
			input.Auth,
			append(
				collectAll(&input.SampleAnalysisParams, &input.FormatParams),
				collect(
					"merge-by-host-and-job", input.MergeByHostAndJob,
					"merge-by-job", input.MergeByJob,
					"clean", input.Clean,
					"last", input.LastN,
				)...,
			),
		)
	})
}

func addSnode(api huma.API) {
	huma.Get(api, "/snode", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"snode",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams),
		)
	})
}

func addSpart(api huma.API) {
	huma.Get(api, "/spart", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			HostAnalysisParams
			FormatParams
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"spart",
			input.Auth,
			collectAll(&input.HostAnalysisParams, &input.FormatParams),
		)
	})
}

func addUptime(api huma.API) {
	huma.Get(api, "/uptime", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			SampleAnalysisParams
			FormatParams
			Interval string `query:"interval"`
			OnlyUp   string `query:"only-up"`
			OnlyDown string `query:"only-down"`
		},
	) (*QueryResponse, error) {
		return queryCommand(
			"uptime",
			input.Auth,
			append(
				collectAll(&input.SampleAnalysisParams, &input.FormatParams),
				collect(
					"interval", input.Interval,
					"only-up", input.OnlyUp,
					"only-down", input.OnlyDown,
				)...,
			),
		)
	})
}

func addVersion(api huma.API) {
	huma.Get(api, "/version", func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
		},
	) (*QueryResponse, error) {
		return queryCommand("version", input.Auth, []string{})
	})
}

// Run a query command.
//
// This must return `error` to be API compatible with Huma, but the error return is always a
// huma.StatusError.

func queryCommand(command, auth string, params []string) (*QueryResponse, error) {
	verbose := Verbose
	if getAuthenticator != nil {
		user, pass := apiutil.DecodeAuth(auth)
		if !getAuthenticator.Authenticate(user, pass) {
			return nil, huma.Error401Unauthorized(command + ": Unknown user/pass combination")
		}
	}
	if verbose && auth != "" {
		Log.Infof("Auth: %q", auth)
	}
	if jobanalyzerDir != "" {
		params = append(params, "-jobanalyzer-dir", jobanalyzerDir)
	}
	if databaseURI != "" {
		params = append(params, "-database-uri", databaseURI)
	}
	// not normally what we want but handy for debugging
	// if verbose {
	// 	params = append(params, "-v")
	// }
	cmdName := "<sonalyze>"
	if verbose {
		Log.Infof(
			"Command: %s %s",
			path.Join(jobanalyzerDir, cmdName),
			command+" "+strings.Join(params, " "),
		)
	}

	anyCmd, _ := cmdlineHandler.ParseVerb(cmdName, command)
	if anyCmd == nil {
		return nil, huma.Error500InternalServerError(command + ": Unknown")
	}
	fs := cmd.NewCLI(command, anyCmd, cmdName, false)
	err := cmdlineHandler.ParseArgs(command, params, anyCmd, fs)
	if err != nil {
		return nil, huma.Error400BadRequest(command + ": " + err.Error())
	}

	// The -cpuprofile option is ignored here, it should have forced ParseArgs to error out.

	var stdoutBuf, stderrBuf strings.Builder
	err = cmdlineHandler.HandleCommand(anyCmd, nil, &stdoutBuf, &stderrBuf)
	// In HandleCommand, the command line parser overrides the global setting.
	Verbose = verbose
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()
	if err != nil {
		return nil, huma.Error400BadRequest(command + ": " + err.Error())
	}
	if stderr != "" {
		Log.Warningf(stderr, "")
	}

	return &QueryResponse{Body: stdout}, nil
}

// Query arguments.
//
// These structures follow the code in ../../cmd/args.go closely, and it's a little sad that this
// code is not there (as a general reification facility).  We should consider whether it can be
// merged in.  But we have our own needs for Huma to think about too.

type Collectable interface {
	Collect() []string
}

type ClusterParams struct {
	Cluster string `query:"cluster"`
}

func (x *ClusterParams) Collect() []string {
	return collect("cluster", x.Cluster)
}

type SourceParams struct {
	FromDate string `query:"from" doc:"ISO time stamp"`
	ToDate   string `query:"to" doc:"ISO time stamp"`
}

func (x *SourceParams) Collect() []string {
	return collect("from", x.FromDate, "to", x.ToDate)
}

type QueryParams struct {
	QueryStmt string `query:"q" doc:"Query expression"`
}

func (x *QueryParams) Collect() []string {
	return collect("q", x.QueryStmt)
}

type HostParams struct {
	// TODO: This is tricky.  The argument parser does not allow comma separation here, but does
	// allow repeated arguments.  Not sure how this works with query args - repeats override, I
	// think?  Do we use the host list parser to separate comma-separated hosts?  If so, we can just
	// forward.  The real problem is if the cli depends on repeated args and we can't do them with
	// the REST API.
	Host string `query:"host" doc:"Comma-separated host ranges"`
}

func (x *HostParams) Collect() []string {
	return collect("host", x.Host)
}

type RecordFilterParams struct {
	HostParams
	User              string `query:"user" doc:"Comma-separated users"`
	ExcludeUser       string `query:"exclude-user" doc:"Comma-separated users"`
	Command           string `query:"command"`
	ExcludeCommand    string `query:"exclude-command"`
	ExcludeSystemJobs string `query:"exclude-system-jobs"`
	Job               string `query:"job"`
	ExcludeJob        string `query:"exclude-job"`
}

func (x *RecordFilterParams) Collect() []string {
	return append(
		x.HostParams.Collect(),
		collect(
			"user", x.User, "exclude-user", x.ExcludeUser,
			"command", x.Command, "exclude-command", x.ExcludeCommand,
			"exclude-system-jobs", x.ExcludeSystemJobs,
			"job", x.Job, "exclude-job", x.ExcludeJob,
		)...,
	)
}

type HostAnalysisParams struct {
	ClusterParams
	SourceParams
	QueryParams
	HostParams
}

func (x *HostAnalysisParams) Collect() []string {
	return collectAll(&x.ClusterParams, &x.SourceParams, &x.QueryParams, &x.HostParams)
}

type SampleAnalysisParams struct {
	ClusterParams
	SourceParams
	QueryParams
	RecordFilterParams
}

func (x *SampleAnalysisParams) Collect() []string {
	return collectAll(&x.ClusterParams, &x.SourceParams, &x.QueryParams, &x.RecordFilterParams)
}

type FormatParams struct {
	Fmt string `query:"fmt" doc:"Format spec"`
}

func (x *FormatParams) Collect() []string {
	return collect("fmt", x.Fmt)
}

type GpuIndexParam struct {
	Gpu string `query:"gpu"`
}

func (x *GpuIndexParam) Collect() []string {
	return collect("gpu", x.Gpu)
}

func collectAll(xs ...Collectable) []string {
	result := make([]string, 0)
	for _, x := range xs {
		result = append(result, x.Collect()...)
	}
	return result
}

func collect(xs ...any) []string {
	if len(xs)%2 == 1 {
		panic("Bad")
	}
	result := make([]string, 0)
	for i := range len(xs) / 2 {
		var s string
		var add bool
		switch v := xs[i*2+1].(type) {
		case string:
			if v != "" {
				s = v
				add = true
			}
		default:
			panic("Type")
		}
		if add {
			result = append(result, "-"+xs[i*2].(string)+"="+s)
		}
	}
	return result
}
