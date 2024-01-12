// Extract job information from the running system and upload it to a server.
//
// TODO: I think we still want heartbeat functionality here?
//
// This is a rethinking of `sonar ps`, without some of the baggage, and integrating the
// functionality of the `exfiltrate` uploader.
//
// Discuss:
//
//   - maybe `jobsampler` is the better name?  `jobprobe`?  `inquisitor`?  `lidar`?
//
// Usage:
//   jobinfo [options] [target]
//
//   Job information data are extracted from the running system and uploaded to the target on JSON
//   format, described below.  A data packet is always written and therefore acts as a heartbeat
//   even if it carries no job records.
//
// Arguments:
//   target
//     The target address for upload.  If provided, this must currently be an http: or https:
//     address.  If omitted, the target is stdout.
//
// Options:
//   -cluster <clustername>
//     Add a "cluster" field to identify the data packet's origin.
//
//   -slurm
//     Read the job ID from Slurm data.  (The default is to synthesize a job ID from the process
//     tree in which a process finds itself.)
//
//   -rollup
//     Merge process records that have the same job ID and command name.
//
//   -exclude-system-jobs
//     Exclude records for system jobs (uid < 1000).
//
//   -exclude-users <user-list>
//     Exclude records whose user names equal any of these comma-separated names.
//
//   -exclude-commands <command-list>
//     Exclude records whose commands start with any of these comma-separated names.
//
//   -min-cpu-time <seconds>
//     Include records for jobs that have used at least this much CPU time.
//
//   -lockdir <directory>
//     Place lockfiles in this directory.  If the lockfile exists on startup, we'll exit
//     immediately.  Note lockfiles should be on a file system that is cleaned on reboot, eg /run.
//
//   -window <seconds>
//     If the target is an http: or https: address, upload data at some random point within
//     a window of <seconds>.  Otherwise this argument is ignored.
//
//   -http-auth <filename>
//     If the target is an http: or https: address, this file will contain user:password for HTTP
//     basic authentication.  Otherwise this argument is ignored.
//
//   -ca-cert <filename>
//     If the target is an https: address, this file will contain the certificate for a CA that allows
//     the client to validate the server.  Otherwise this argument is ignored.
//
//   -nvidia
//     Attempt to obtain NVIDIA GPU data by running nvidia-smi or probing devices.  It's usually safe
//     to do this even on systems without an NVIDIA GPU.
//
//   -amd
//     Attempt to obtain AMD GPU data by running rocm-smi or probing devices.  It's usually safe
//     to do this even on systems without an AMD GPU.
//
// The data packet is a JSON object with these fields:
//
//   Since v=0.1.0:
//
//   v       - string  - data semver
//   time    - integer - seconds since unix epoch UTC
//   host    - string  - host name
//   cluster - string  - cluster name, as presented on the command line, if any
//   jobs    - array   - job objects
//
// Each job object is a JSON object with these fields:
//
//   Since v=0.1.0:
//
//   user       - string  - user name for entry
//   job        - integer - job number
//   pid        - integer - process ID
//   cmd        - string  - command string
//   cpu_pct    - float   - cpu time ...
//   cpu_s      - integer - cpu time ...
//   vmem_kb    - integer - virtual memory size in KiB
//   pss_kb     - integer - resident (proportional set size) memory size in KiB
//   gpus       - string  - gpu set: "none", "unknown", "n,m,o,..." (default "none")
//   gpu_pct    - float   - gpu time ...
//   gpumem_kb  - integer - gpu resident memory size in KiB
//   gpumem_pct - float   - gpu resident memory size as percent of total
//   gpufail    - integer - 0 means OK, the rest are in flux
//   rolledup   - integer - number of jobs rolled up
//
// Semantics are as for sonar, ie, complicated.  DOCUMENTME.
//
// Job object fields that have their default values are omitted (zeroes, empty strings, or for the
// gpu set, "none").

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"go-utils/auth"
	"go-utils/filesys"
	"go-utils/httpclient"
	"go-utils/status"
	"go-utils/sysinfo"
)

const (
	dataVersion       = "0.1.0"
	maxAttempts       = 6
	resendIntervalMin = 5
)

var (
	// Command line variables, cleaned up
	target            *url.URL
	clusterName       string
	slurm             bool
	rollup            bool
	excludeSystemJobs bool
	excludeCommands   []string
	excludeUsers      []string
	minCpuTime        uint
	lockDir           string
	window            uint
	httpUser          string
	httpPassword      string
	caCertFile        string
	verbose           bool
	nvidia            bool
	amd               bool
)

var (
	// Global system information
	memTotalKiB         uint64
	pageSizeKiB         uint
	bootTimeSinceEpochS int64
	clockTicksPerS      uint
	hostName            string
	nowSinceEpochS      int64
)

type envelopeObject struct {
	V       string       `json:"v"`
	Time    int64        `json:"time"`
	Cluster string       `json:"cluster,omitempty"`
	Host    string       `json:"host"`
	Jobs    []*jobObject `json:"jobs"`
}

type jobObject struct {
	User      string  `json:"user"`
	Job       int     `json:"job,omitempty"`
	Pid       int     `json:"pid,omitempty"`
	Cmd       string  `json:"cmd"`
	CpuPct    float64 `json:"cpu_pct,omitempty"`
	CpuSec    int64   `json:"cpu_s,omitempty"`
	VmemKb    int64   `json:"vmem_kb,omitempty"`
	PssKb     int64   `json:"pss_kb,omitempty"`
	Gpus      string  `json:"gpus,omitempty"`
	GpuPct    float64 `json:"gpu_pct,omitempty"`
	GpumemKb  int64   `json:"gpumem_kb,omitempty"`
	GpumemPct float64 `json:"gpumem_pct,omitempty"`
	GpuFail   int     `json:"gpufail,omitempty"`
	Rolledup  int     `json:"rolledup,omitempty"`
}

func main() {
	status.Start("jobanalyzer/jobinfo")
	parseCommandLine()

	var lf *filesys.Lockfile
	if lockDir != "" {
		lf = filesys.NewLockfile(lockDir, "jobinfo-lock."+hostName)
		if lf == nil {
			return
		}
	}

	err := getSystemInformation()
	if err != nil {
		status.Fatalf("Failed to obtain system information, %v", err)
	}

	// Extract job information
	jobs := getJobs()

	// Unlock now, since the upload phase may linger if the target address is temporarily not
	// reachable.
	if lf != nil {
		lf.Unlock()
	}

	// Create the output and send it to the right place
	var e envelopeObject
	e.V = dataVersion
	e.Time = nowSinceEpochS
	e.Cluster = clusterName
	e.Host = hostName
	e.Jobs = jobs
	dataBytes, err := json.Marshal(&e)
	if err != nil {
		status.Fatalf("Could not format data as json: %v", err)
	}
	if target == nil {
		_, err := os.Stdout.Write(dataBytes)
		fmt.Println()
		if err != nil {
			status.Fatalf("Failed to write all data, %v", err)
		}
	} else {
		if window > 0 {
			secs := rand.Intn(int(window))
			if verbose {
				fmt.Printf("Sleeping %d seconds\n", secs)
			}
			time.Sleep(time.Duration(secs) * time.Second)
		}
		client, err := httpclient.NewClient(
			target,
			caCertFile,
			httpUser,
			httpPassword,
			maxAttempts,
			resendIntervalMin,
			verbose,
		)
		if err != nil {
			status.Fatalf("Failed to create HTTP client, %v", err)
		}
		client.PostDataByHttp("", dataBytes)
		// ProcessRetries will return when all packets have been sent or we've timed out
		client.ProcessRetries()
	}
}

func getSystemInformation() (err error) {
	// Installed memory is needed for all sorts of things
	m, err := sysinfo.PhysicalMemoryBy()
	if err != nil {
		return
	}
	memTotalKiB = m / 1024

	// We need this for boot-relative times in /proc
	bootTimeSinceEpochS, err = sysinfo.BootTime()
	if err != nil {
		return
	}

	// Some data are presented in pages
	pageSizeKiB = sysinfo.PagesizeBy() / 1024

	// Some elapsed times are presented in terms of ticks.  On Linux, a tick is a constant 1/100s,
	// and though it is possible to do sysconf(_SC_CLK_TCK) it is evidently not necessary.
	//
	// Quoting from https://github.com/tklauser/go-sysconf/blob/main/sysconf_linux.go,
	//
	//   CLK_TCK is a constant on Linux for all architectures except alpha and ia64.
	//   See e.g.
	//   https://git.musl-libc.org/cgit/musl/tree/src/conf/sysconf.c#n30
	//   https://github.com/containerd/cgroups/pull/12
	//   https://lore.kernel.org/lkml/agtlq6$iht$1@penguin.transmeta.com/
	//
	// The last one of those is probably most interesting.  Quoting Linus:
	//
	//   The fact that libproc believes that HZ can change is _their_ problem.
	//   I've told people over and over that user-level HZ is a constant (and, on
	//   x86, that constant is 100), and that won't change.
	clockTicksPerS = 100

	// Node name is needed for all readings
	hostName, err = os.Hostname()
	if err != nil {
		return
	}

	// Everyone uses the same timestamp
	nowSinceEpochS = time.Now().UTC().Unix()

	return
}

// Command line syntax errors are reported to flag's error output + exit(2).
// Other errors are logged to the syslog.

func parseCommandLine() {
	errout := flag.CommandLine.Output()
	flag.Usage = func() {
		fmt.Fprintf(errout, `
Usage of %s:
  %s [options] [target-address]

Arguments:
  target-address is a network endpoint for upload, typically https://host:port/service;
  default stdout.

Options:
`,
			os.Args[0],
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.StringVar(&clusterName, "cluster", "",
		"Add a \"cluster\" field with `cluster-name` to identify the record")
	flag.BoolVar(&slurm, "slurm", false,
		"Get the job ID from slurm data")
	flag.BoolVar(&rollup, "rollup", false,
		"Merge process records that have the same job ID and command name")
	flag.BoolVar(&excludeSystemJobs, "exclude-system-jobs", false,
		"Exclude records for system jobs (uid < 1000)")
	var excludeCmdList string
	flag.StringVar(&excludeCmdList, "exclude-commands", "",
		"Exclude records whose commands start with a string in `command,...`")
	var excludeUserList string
	flag.StringVar(&excludeUserList, "exclude-users", "",
		"Exclude records whose users match strings in `user,...`")
	flag.UintVar(&minCpuTime, "min-cpu-time", 0,
		"Include only records for jobs that have used at least `seconds` CPU time")
	flag.StringVar(&lockDir, "lockdir", "",
		"Use `directory` to hold lock files (default is no locking)")
	flag.UintVar(&window, "window", 0,
		"Upload window in `seconds`, for network targets")
	var httpAuthFile string
	flag.StringVar(&httpAuthFile, "http-auth", "",
		"File with username:password for basic HTTP authentication, for HTTP/S targets")
	flag.StringVar(&caCertFile, "ca-cert", "",
		"File with HTTPS server certificate, for HTTPS targets")
	flag.BoolVar(&nvidia, "nvidia", false,
		"Probe for NVIDIA device data")
	flag.BoolVar(&nvidia, "amd", false,
		"Probe for AMD device data")
	flag.BoolVar(&verbose, "v", false,
		"Verbose output")
	flag.Parse()
	rest := flag.Args()
	if len(rest) > 1 {
		flag.Usage()
		os.Exit(2)
	}
	targetArg := ""
	if len(rest) > 0 {
		targetArg = rest[0]
	}

	var isHttp, isHttps bool
	if targetArg != "" {
		// TODO: target validation.  The URL parser seems to accept pretty much anything.  Probably we
		// require scheme://host:port and no path on the host and no query.
		var err error
		target, err = url.Parse(targetArg)
		if err != nil || target.Scheme == "" || target.Host == "" || target.Path != "" {
			errmsg := ""
			if err != nil {
				errmsg = fmt.Sprintf(": %v", err)
			}
			fmt.Fprintf(errout, "Failed to parse target URL %s%s\n", target, errmsg)
			os.Exit(2)
		}

		isHttp = target.Scheme == "http" || target.Scheme == "https"
		isHttps = target.Scheme == "https"
		if !isHttp {
			fmt.Fprintf(errout, "Only http / https targets for now\n")
			os.Exit(2)
		}
	}

	if isHttps && caCertFile == "" {
		fmt.Fprintf(errout, "Missing server certificate argument (-ca-cert) for https target")
		os.Exit(2)
	}

	excludeCommands = strings.Split(excludeCmdList, ",")
	excludeUsers = strings.Split(excludeUserList, ",")

	if lockDir != "" {
		lockDir = path.Clean(lockDir)
		info, err := os.DirFS(lockDir).(fs.StatFS).Stat(".")
		if err != nil || !info.IsDir() {
			status.Fatalf("Failed to stat -lockdir directory %s", lockDir)
		}
	}

	if isHttp && httpAuthFile != "" {
		var err error
		httpUser, httpPassword, err = auth.ParseAuth(httpAuthFile)
		if err != nil {
			status.Fatalf("Failed to read -http-auth file %s", httpAuthFile)
		}
	}
}
