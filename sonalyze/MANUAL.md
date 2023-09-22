# `sonalyze` manual

## USAGE

Analyze `sonar` log files and print information about jobs or systems.

### Summary

```
sonalyze operation [options] [-- logfile ...]
```

where `operation` is `jobs`, `load`, or `help`.

The `jobs` operation prints information about jobs, collected from the sonar records.

The `load` operation prints information about the load on the systems, collected from the sonar
records.

The `help` operation prints high-level usage information.

Run `sonalyze <operation> help` to get help about options for the specific operation.

### Overall operation

The program operates by phases:

* reading any system configuration files
* computing a set of input log files
* reading these log files with input filters applied, resulting in a set of input records
* aggregating data across the input records
* filtering the aggregated data with the aggregation filters
* printing the aggregated data with the output filters

Input filtering options are shared between the operations.  Aggregation filtering and output options
are per-operation, as outlined directly below.

### Log file computation options

`--data-path=<path>`

  Root directory for log files, overrides the default.  The default is the `SONAR_ROOT` environment
  variable, or if that is not defined, `$HOME/sonar_logs`.

`-- <filename>`

  If present, each `filename` is used for input instead of anything reachable from the data path;
  the data path is ignored.

### System configuration options

`--config-file=<path>`

  Read a JSON file holding system information keyed by hostname.  This file is required by options
  or print formats that make use of system-relative values (such as `rcpu`).  See the section
  "SYSTEM CONFIGURATION FILES" below.

### Input filter options

All filters are optional.  Records must pass all specified filters.

`-u <username>`, `--user=<username>`

  The user name(s), the option can be repeated.  The default for `jobs` is the current user,
  `$LOGNAME`, except in the case of `--job=`, `--zombie`, or `--exclude-user=`, when the default is
  everyone.  The default for `load` is everyone.  Use `-` to ask for everyone.

`--exclude-user=<username>`

  Normally, users `root` and `zabbix` are excluded from the report.  (They don't run jobs usually,
  but with synthesized jobs they can appear in the log anyway.)  With the exclude option, list
  *additional* user names to be excluded.  The option can be repeated

`--command=<command>`

  Select only records whose command name matches `<command>` exactly.  This option can be repeated.

`--exclude-command=<command>`

  Exclude commands matching `<command>` exactly.  This option can be repeated.

`-j <job#>`, `--job=<job#>`

  Select specific records by job number(s).  The option can be repeated.

`-f <fromtime>`, `--from=<fromtime>`

  Select only records with this time stamp and later, format is either `YYYY-MM-DD`, `Nd` (N days ago)
  or `Nw` (N weeks ago).  The default is `1d`: 24 hours ago.

`-t <totime>`, `--to=<totime>`

  Select only records with this time stamp and earlier, format is either `YYYY-MM-DD`, `Nd` (N days
  ago) or `Nw` (N weeks ago).  The default is now.

`--host=<hostname>`

  Select only records from these host names.  The host name filter applies both to file name
  filtering in the data path and to record filtering within all files processed (as all records also
  contain the host name).  The default is all hosts.  The host name can use wildcards and expansions
  in some ways; see later section.  The option can be repeated.

### Job filtering and aggregation options

These are only available with the `jobs` command.  All filters are optional.  Jobs must pass all
specified filters.

`-b`, `--batch`

  Aggregate data across hosts (this would normally be appropriate for systems with a batch queue,
  such as Fox).

`--min-cpu-avg=<pct>`, `--max-cpu-avg=<pct>`

  Select only jobs that have at least / at most `pct` percent (an integer, one full CPU=100) average
  CPU utilization.

`--min-cpu-peak=<pct>`, `--max-cpu-peak=<pct>`

  Select only jobs that have at least / at most `pct` percent (an integer, one full CPU=100) peak
  CPU utilization.

`--min-rcpu-avg=<pct>`, `--max-rcpu-avg=<pct>`, `--min-rcpu-peak=<pct>`, `--max-rcpu-peak=<pct>`

  Select only jobs that have at least / at most `pct` percent (an integer, the entire system=100)
  average or peak system-relative CPU utilization.  Requires a system config file.

`--min-mem-avg=<size>`

  Select only jobs that have at least `size` gigabyte average main memory utilization.

`--min-mem-peak=<size>`

  Select only jobs that have at least `size` gigabyte peak main memory utilization.

`--min-rmem-avg=<pct>`, `--min-rmem-peak=<pct>`

  Select only jobs that have at least `pct` percent (an integer, the entire system=100) average or
  peak main memory utilization.  Requires a system config file.

`--min-gpu-avg=<pct>`, `--max-gpu-avg=<pct>`

  Select only jobs that have at least / at most `pct` percent (an integer, one full device=100)
  average GPU utilization.  Note that most programs use no more than one accelerator card, and there
  are fewer of these than CPUs, so this number will be below 100 for most jobs.
   
`--min-gpu-peak=<pct>`, `--max-gpu-peak=<pct>`

  Select only jobs that have at least / at most `pct` percent (an integer, one full device=100) peak
  GPU utilization.

`--min-rgpu-avg=<pct>`, `--max-rgpu-avg=<pct>`, `--min-rgpu-peak=<pct>`, `--max-rgpu-peak=<pct>`

  Select only jobs that have at least / at most `pct` percent (an integer, the entire system=100)
  average or peak system-relative GPU utilization.  Requies a system config file.
  
`--min-gpumem-avg=<pct>`

  Select only jobs that have at least `pct` percent (an integer, one full device=100) average GPU
  memory utilization.

`--min-gpumem-peak=<pct>`

  Select only jobs that have at least `pct` percent (an integer, one full device=100) peak GPU
  memory utilization.

`--min-rgpumem-avg=<pct>`, `--min-rgpumem-peak=<pct>`

  Select only jobs that have at least `pct` percent (an integer, the entire system=100) average or
  peak GPU memory utilization.  Requires a system config file.

`--min-runtime=<time>`

  Select only jobs that ran for at least the given amount of time.  Time is given on the formats
  `WwDdHhMm` where the `w`, `d`, `h`, and `m` are literal and `W`, `D`, `H`, and `M` are nonnegative
  integers, all four parts -- weeks, days, hours, and minutes -- are optional but at least one must
  be present.  (Currently the parts can be in any order but that may change.)

`--no-gpu`

  Select only jobs that did not use any GPU.

`--some-gpu`

  Select only jobs that did use some GPU (even if the GPU avg/max statistics round to zero).

`--completed`

  Select only jobs that have completed (have no samples at the last time recorded in the log).

`--running`

  Select only jobs that are still running (have a sample at the last time recorded in the log).

`--zombie`

  Select only jobs deemed to be zombie jobs.  (This includes actual zombies and defunct processes.)

`--min-samples`

  Select only jobs with at least this many samples.  (There may be multiple samples at the same
  time instant for a single job if the job has multiple processes with different names, so this
  option does not guarantee that a job is observed at different points in time.  Use `--min-runtime`
  if that's what you mean.)

### Load filtering and aggregation options

These are only available with the `load` command.  All filters are optional.  Records must pass all
specified filters.

`--hourly`

  Bucket the records hourly and present averages (the default).  Contrast `--daily` and `--none`.

`--daily`

  Bucket the records daily and present averages.  Contrast `--hourly` and `--none`.

`--none`

  Do not bucket the records.  Contrast `--hourly` and `--daily`.

### Job printing options

`--breakdown=<keywords>`

  For a job, also print a breakdown according to the `<keywords>`.  The keywords are `host` and
  `command` and can be present in either order.  Suppose jobs are aggregated across hosts (with
  `--batch`) and that the jobs may run as multiple processes with different names.  Adding
  `--breakdown=host,command` will show the summary for the job, but then break it down by host, and
  for each host, break it down by command, showing a summary line per host (across all the commands
  on that host) and then a line for each command.  This yields insight into how the different
  commands contribute to the resource use of the job, and how the jobs balance across the different
  hosts.
  
  To make the printout comprehensible, the first field value of each first-level breakdown lines is
  prefixed by `*` and the first field value of each second-level breakdown line is prefixed by `**`
  (in both plain text and csv output forms).  Any consumer must be prepared to handle this, should
  it be exposed to this type of output.
  
`-n <number-of-jobs>`, `--numjobs=<number-of-jobs>`

  Show only the *last* `number-of-jobs` selected jobs per user.  The default is "all".  Selected
  jobs are sorted ascending by the start time of the job, so this option will select the last
  started jobs.

`--fmt=<format>`

  Format the output for `load` according to `format`, which is a comma-separated list of keywords,
  see OUTPUT FORMAT below.


### Load printing options

The *absolute load* at an instant on a host is the sum of a utilization field across all the
records for the host at that instant, for the cpu, memory, gpu, and gpu memory utilization.  For
example, on a system with 192 cores the maximum absolute CPU load is 19200 (because the CPU load
is a percentage of a core) and if the system has 128GB of RAM then the maximum absolute memory
load is 128.
  
The absolute load for a time interval is the average for each of those fields across all the
absolute loads in the interval.

The *relative load* is the absolute load of a system (whether at an instance or across an interval)
relative to the host's configuration for the quantity in question, as a percentage.  If the absolute
CPU load at some instant is 5800 and the system has 192 cores then the relative CPU load at that
instant is 5800/19200, ie 30%.

`--last`

  Print only records for the last instant in time (after filtering/bucketing).  Contrast `--all`.

`--all`

  Print the records for all instants in time (after filtering/bucketing).  Contrast `--last`.

`--fmt=<format>`

  Format the output for `load` according to `format`, which is a comma-separated list of keywords,
  see OUTPUT FORMAT below.

`--compact`

  Do not print any output for empty time slots.  The default is to print a record for every time
  slot in the requested range.


## MISC EXAMPLES

Many examples of usage are with the use cases in [../README.md](../README.md).  Here are some more:

List all my jobs the last 24 hours:

```
sonalyze jobs
```

List the jobs for all users from up to 2 weeks ago in the given log file (presumably containing data
for the entire time period) that used at least 10 cores worth of CPU on average and no GPU:

```
sonalyze jobs --user=- --from=2w --min-cpu-avg=1000 --no-gpu -- ml8.hpc.uio.no.csv
```

## LOG FILES

The log files under the log root directory -- ie when log file names are not provided on the command
line -- are expected to be in a directory tree coded first by four-digit year (CE), then by month
(01-12), then by day (01-31), with a file name that is the name of a host with the ".csv" extension.
That is, `$SONAR_ROOT/2023/06/26/deathstar.hpc.uio.no.csv` could be such a file.

## HOST NAME PATTERNS

A host name *pattern* specifies a set of host names.  The pattern consists of literal characters,
range expansions, and suffix wildcards.  Consider `ml[1-4,8]*.hpc*.uio.no`.  This matches
`ml1.hpc.uio.no`, `ml1x.hpcy.uio.no`, and several others.  In brief, the host name is broken into
elements at the `.`.  Then each element can end with `*` to indicate that we match a prefix of the
input.  The values in brackets are expanded: Ranges m-n turn into m, m+1, m+2, ..., n (inclusive),
stand-alone values stand for themselves.

The pattern can have fewer elements than the host names we match against, typically the unqualified host
name is used: `--host ml[1-4,8]` will select ML nodes 1, 2, 3, 4, and 8.

## SYSTEM CONFIGURATION FILES

The system configuration files are JSON files providing the details for each host.  See
[`ml-nodes.json`](../production/ml-nodes/ml-nodes.json) for an example.

The format is an array containing objects `[{...}, {...}, ...]`, each object describing
one host or node.  The fields are:

* `hostname`: the fully qualified domain name for the host
* `description`: human-readable text summarizing CPUs, RAM, GPUS, and VRAM
* `cpu_cores`: integer, the total number accounting for hyperthreads too
* `mem_gb`: integer
* `gpu_cards`: integer
* `gpumem_gb`: integer, the total amount of memory across all cards

There's an assumption here that all CPUs and GPUs on a system are of the same type.  More fields can
easily be added.

The description should be useful, because sometimes it's the only datum that's shown to a user, eg:
```
"2x14 Intel Xeon Gold 5120 (hyperthreaded), 128GB, 4x NVIDIA RTX 2080 Ti @ 11GB"
```

## OUTPUT FORMAT

The `--fmt` switch controls the format for the command through a list of keywords.  Each keyword
adds a column to the output.  In addition to the keywords that are command-specific (and listed
below) there are some general ones:

* `csv` forces CSV-format output, the default is fixed-column layout
* `csvnamed` forces CSV-format output with each field prefixed by `<fieldname>=`
* `header` forces a header to be printed, default for fixed-column
* `noheader` forces a header not to be printed, default for `csv` and `csvnamed`
* `tag:something` forces a field `tag` to be printed for each record with the value `something`

### Jobs

Output records are sorted in order of increasing start time of the job.

The formatting keywords for the `jobs` command are

* `now` is the current time on the format `YYYY-MM-DD HH:MM`
* `job` is a number
* `jobm` is a number, possibly suffixed by a mark "!" (job is running at the start and end of the time interval),
  "<" (job is running at the start of the interval), ">" (job is running at the end of the interval).
* `user` is the user name
* `duration` on the format DDdHHhMMm shows the number of days DD, hours HH and minutes MM the job ran for.
* `start` and `end` on the format `YYYY-MM-DD HH:MM` are the endpoints for the job
* `cpu-avg`, `cpu-peak`, `gpu-avg`, `gpu-peak` show CPU and GPU utilization as
   percentages, where 100 corresponds to one full core or device, ie on a system with 64 CPUs the
   CPU utilization can reach 6400 and on a system with 8 accelerators the GPU utilization can reach 800.
* `mem-avg`, `mem-peak`, `gpumem-avg`, and `gpumem-peak` show main and GPU memory average and peak
   utilization in GB
* `rcpu-avg`, ..., `rmem-avg`, ... are available to show relative usage (percentage of full system capacity).
   These require a config file for the system to be provided with the `--config-file` flag.
* `gpus` is a comma-separated list of device numbers used by the job
* `host` is a list of the host name(s) running the job (showing only the first element of the FQDN, and 
  compressed using the same patterns as in HOST NAME PATTERNS above)
* `cmd` is the command name, as far as is known.  For jobs with multiple processes that have different
   command names, all command names are printed.
* `cpu` is an abbreviation for `cpu-avg,cpu-peak`, `mem` an abbreviation for `mem-avg,mem-peak`, and so on,
  for `gpu`, `gpumem`, `rcpu`, `rmem`, `rgpu`, and `rgpumem`
* `std` is an abbreviation for `jobm,user,duration,host`

The default keyword set is `std,cpu,mem,gpu,gpumem,cmd`.

### Systems

Output records are sorted in ...

The host name is printed on a separate line before the data for each host.

The formatting keywords for the `load` command are as follows, all fields pertain to the records or summary
records, except `now`:

* `date` (`YYYY-MM-DD`)
* `time` (`HH:MM`)
* `cpu` (percentage, 100=1 core)
* `rcpu` (percentage, 100=all system cores)
* `mem` (GB)
* `rmem` (percentage, 100=all system memory)
* `gpu` (percentage, 100=1 card)
* `rgpu` (percentage, 100=all cards)
* `gpumem` (GB)
* `rgpumem` (percentage, 100=all memory on all cards)
* `gpus` (list of GPUs)
* `now` is the current time on the format `YYYY-MM-DD HH:MM`

The default keyword set is `date,time,cpu,mem,gpu,gpumem,gpus`.
