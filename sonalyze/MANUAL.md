# `sonalyze` manual

## USAGE

Analyze `sonar` log files and print information about jobs or systems.

### Summary

```
sonalyze operation [options] [-- logfile ...]
```

where `operation` is `jobs`, `load`, `uptime`, `profile`, `version`, `parse`, `metadata`, or `help`:

* The `jobs` operation prints information about jobs, collected from the sonar records.
* The `load` operation prints information about the load on the systems, collected from the sonar
  records.
* The `uptime` operation prints information about when systems and their components were up
  or down, collected from sonar records.
* The `profile` operation prints information about the behavior of a single job across time.
* The `parse` and `metadata` operations are for testing, mainly: They perform low-level operations on
  the sonar logs and print the results.
* The `version` operation prints machine-readable information about sonalyze and its configuration,
  useful mostly for testing.
* The `help` operation prints high-level usage information.

Run `sonalyze <operation> help` to get help about options for the specific operation.

All the operations take a `--fmt` option to control the output; run with `--fmt=help` to get help on
formatting options.

### Overall operation

The program operates by phases:

* reading any system configuration files
* computing a set of input log files
* reading these log files with record filters applied, resulting in a set of input records
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

### Record filter options

All filters are optional.  Records must pass all specified filters.  All commands support the record
filters.

`-u <username>`, `--user=<username>`

  The user name(s), the option can be repeated.  Use `-` to ask for everyone.

  The default value for `--user` is the the current user, `$LOGNAME`, but there are several
  exceptions.  When the command is `load`, or when the `--job` and `--exclude-user` record filters
  are used, or when the `--zombie` job filter is being used then the default is "everyone".

  Additionally, when the `--job` record filter is used, then users "root" and "zabbix" are excluded.

`--exclude-user=<username>`

  Request that records with these user names are excluded, in addition to normal exclusions.
  The option can be repeated.  See above about defaults.

`--command=<command>`

  Select only records whose command name matches `<command>` exactly.  This option can be repeated.

  The default value for `--command` is every command for `parse`, `metadata` and `uptime`, and every
  command except some system commands for `jobs` and `load` (currently `bash`, `zsh`, `sshd`,
  `tmux`, and `systemd`; this is pretty ad-hoc).

`--exclude-command=<command>`

  Exclude commands matching `<command>` exactly, in addition to default exclusions.  This option can
  be repeated.

`-j <job#>`, `--job=<job#>`

  Select specific records by job number(s).  The option can be repeated.

`-f <fromtime>`, `--from=<fromtime>`

  Select only records with this time stamp and later, format is either `YYYY-MM-DD`, `Nd` (N days ago)
  or `Nw` (N weeks ago).  The default is `1d`: 24 hours ago.

`-t <totime>`, `--to=<totime>`

  Select only records with this time stamp and earlier, format is either `YYYY-MM-DD`, `Nd` (N days
  ago) or `Nw` (N weeks ago).  The default is `0d`: now.

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

### Uptime printing options

`--interval=<interval>`

  This required argument specifies to the sampling interval used by sonar in minutes, though it is
  currently useful to choose an interval about 1 minute shorter than sonar here.  It is used to
  determine whether there are any gaps in the system timeline, indicating system downtime.

`--only-down`

  Show only records for when the system or its components were down.  The default is to show all records.

`--only-up`

  Show only records for when the system or its components were up.  The default is to show all records.

`--fmt=<format>`

  Format the output for `uptime` according to `format`, which is a comma-separated list of keywords,
  see OUTPUT FORMAT below.

### Profiling printing options

`--fmt=<format>`

  Format the output for `profile` according to `format`, which is a comma-separated list of keywords,
  see OUTPUT FORMAT below.

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

List all the processes in the given job for the last 24 hours, broken down at each sampler step:

```
sonalyze profile -j 12345
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

The `--fmt` switch controls the format for the command output through a list of field names and
control keywords.  Each field name adds a column to the output; field names are specific to the
command.  In addition to the field names there are control keywords that determine the output
format:

* `fixed` forces human-readable output in fixed-width columns with a header
* `csv` forces CSV-format output, the default is fixed-column layout
* `csvnamed` forces CSV-format output with each field prefixed by `<fieldname>=`
* `json` forces JSON-format output (without any header, ever)
* `header` forces a header to be printed for CSV; this is the default for fixed-column output
   (and a no-op for JSON)
* `noheader` forces a header not to be printed, default for `csv`, `csvnamed`, and `json`
* `nodefaults` applies in some cases with some output forms (notably the `parse` command with
  csv or JSON output) and causes the suppression of fields that have their default values)
* `tag:something` forces a field `tag` to be printed for each record with the value `something`

Run with `--fmt=help` to get help on formatting syntax, field names, aliases, and controls, as
well as a description of the default fields printed.

### Field names for `jobs`

Output records are sorted in order of increasing start time of the job.

The default format is `fixed`.  The field names for the `jobs` command are at least these:

* `now` is the current time on the format `YYYY-MM-DD HH:MM`
* `now/sec` is the current time as a unix timestamp
* `job` is a number
* `jobm` is a number, possibly suffixed by a mark "!" (job is running at the start and end of the time interval),
  "<" (job is running at the start of the interval), ">" (job is running at the end of the interval).
* `user` is the user name
* `duration` on the format DDdHHhMMm shows the number of days DD, hours HH and minutes MM the job ran for.
* `duration/sec` is the duration of the job (in real time) in seconds
* `cputime` on the format DDdHHhMMm shows the number of CPU days DD, hours HH and minutes MM the job
   ran for, note this is frequently longer than `duration`
* `cputime/sec` is the total CPU time of the job in seconds, across all cores
* `gputime` on the format DDdHHhMMm shows the number of GPU days DD, hours HH and minutes MM the job
   ran for, note this too can be longer than `duration`, if a computation used multiple cards in parallel
* `gputime/sec` is the total GPU time of the job in seconds, across all cards
* `start` and `end` on the format `YYYY-MM-DD HH:MM` are the endpoints for the job
* `start/sec` and `end/sec` are unix timestamps for the endpoints of the job
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
* `std` is an abbreviation for the set of default fields

A Unix timestamp is the number of seconds since 197-01-01T00:00:00UTC.

### Field names for `load`

Output records are sorted in TBD order.

The host name is printed on a separate line before the data for each host.

The default format is `fixed`.  The field for the `load` command are as follows:

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

### Field names for `parse`, `metadata` and `uptime`

Consult the on-line help for details.

Note that `parse` has an alias `roundtrip` that causes it to print the selected records with a
format that makes them suitable for input to sonalyze.

## COOKBOOK: CRUDE, HIGH-LEVEL JOB PROFILING

This is based on an actual use case.  Suppose you have a pipeline or complex job of some sort that's
failing.  In the simplest case, the sonar logs are already there, so can you examine them?

To do this, figure out the job number and then run

```
$ sonalyze profile -j <job-number>
```

(If the profile was not obtained during the last 24 hours, remember to add a `--from` argument too!)

This will print out summaries of all the samples for the job across its lifetime (there may be a
significant amount of output).  If the job had multiple commands or processes, each will be listed
separately for each time step.  (In the future, if the job ran on multiple hosts, it will also be
broken down across hosts.)

In the less simple case, the sonar logs are not available because sonar was not running (typical for
a cloud platform or VM, perhaps) or sonar ran with insufficient resolution to allow the problem to be
diagnosed.

In this case, we run sonar in the background while the job is running to collect appropriate
samples, here the interval (`-i`) is 30s instead of the sonar default of 5m and we tell the script where
to find sonar and where to store the output:

```
$ .../Jobanalyzer/sonard/sonard -i 30 -s .../sonar/target/release/sonar my-logfile.csv &
```

While that is running, run the job:
```
$ my_pipeline
```

and when the job is done, stop sonar again:
```
$ pkill sonard
```

Now analyze that log:
```
$ sonalyze profile -j <job-number> -- my-logfile.csv
```

Sonar runs in about 100ms; the smallest sensible interval is probably in the vicinity of 5s, but in
that case there will be a vast amount of output for all but the shortest runs.

### Plotting data

With `sonalyze profile` it's possible to generate JSON, CSV, or HTML data.  The JSON data contain
all the data points, while CSV and HTML data can show only one quantity at a time per process and
time step.  For example,

```
$ sonalyze profile -j <job-number> --fmt=csv,cpu -- my-logfile.csv
```

prints the CPU values from the profile.  The output looks like this:

```
time,bwa (1119426),samtools (1119428)
2023-10-21 08:35,3075,
2023-10-21 08:40,3094,
2023-10-21 08:45,3092,11
2023-10-21 08:50,3078,12
2023-10-21 08:55,3093,11
2023-10-21 09:00,3080,14
2023-10-21 09:05,3091,11
2023-10-21 09:10,,10
2023-10-21 09:15,,13
```

That is, the first line is a heading listing the constituent processes, and each successive line has the
time stamp and the data for each process.  If a process was not alive at a given time the field is present,
but empty.  Spreadsheets can ingest this data format and plot it.

It is perhaps more useful to generate HTML, which can be viewed in a web browser:
```
$ sonalyze profile -j <job-number> --fmt=html,cpu -- my-logfile.csv > my-profile.html
```

Open the file in a browser; it renders the profile graphically.

### Dealing with outliers

Sometimes the profiles contain extreme outliers resulting from calculation errors, sampling
conincidences, and so on.  In those cases, the profiles may be hard to understand.  A useful switch
is `--max`, which will clamp values to the given maximum value (or to zero, if the values are more
than twice the max, on the assumption that these data are pure noise).  Suppose you know you have no
more than 64 cores on the machine.  Then this is meaningful (remember CPU utilization is presented
in percent of one core):

```
$ sonalyze profile -j <job-number> --fmt=html,cpu --max=6400 -- my-logfile.csv > my-profile.html
```

### Dealing with spikes and with large data

Some data are spiky: the process uses a modest amount of resources most of the time but has spikes
when it uses much more.  Plotting the raw data, especially for data that came from very frequent
sampling, may hide the average behavior in the spikes because the spikes are prominent in the plot.
Smoothing the data has the double effect of exposing the averages by removing the spikes, and
reducing the data volume.  Averages over 2-10 points seem to work well, we use the `--bucket` switch:

```
$ sonalyze profile -j <job-number> --fmt=html,cpu --bucket=3 -- my-logfile.csv > my-profile.html
```

Bucketing is performed after clamping.
