# `sonalyze` manual

## USAGE

`sonalyze` is the database manager and query interface for time series data coming from sonar,
sacctd, and other monitoring components.  Use it to add data and to query, aggregate, and export
data.

There's a quick introduction to how everything hangs together in `doc/HOWTO.md` at the top level of
the repo.

### Data model at a glance

Notionally there are these tables:

* `cluster` is the table of clusters known to the DB manager, with their aliases and descriptions
* `config` is the table of per-node configuration information
* `sysinfo` is the table of per-node system information extracted by Sonar on each node every day
* `sacct` is the table of completed Slurm jobs on the cluster
* `sample` is the table of Sonar samples, collected by Sonar on each node

The sonalyze operations (next section) insert data into a table, extract raw data from a table, or
generate cooked data from one or more tables.

See db/README.md for more information about the data model and how it maps onto the data store.

### Summary of operations

```
sonalyze operation [options] [-- logfile ...]
```

where `operation` is one of the following.

There is one data insertion operation that works on the `sysinfo`, `sacct`, and `sample` tables:

* The `add` operation appends new records to the database (the database is append-only).

There are raw data extractors that work on a single table:

* The `cluster` operation prints information from the `cluster` table
* The `config` operation prints information from the `config` table
* The `sacct` operation prints information from the `sacct` table
* The `sample` operation prints information from the `sample` table
* The `node` operation prints information from the `sysinfo` table (this operation should have
  been called `sysinfo` but that name is currently taken)

Then there are built-in aggregation operations that work on the `sample` table joined with the
`config` table.  You almost never want raw sample data, but cooked data from some of the aggregation
operations instead.  The data thus requested could have been represented in (precomputed or
lazily-computed) tables, but at the moment they are not and this of course affects performance as
the data have to be computed when they are requested:

* The `jobs` operation prints information about jobs
* The `load` operation prints information about system load
* The `uptime` operation prints information about system and component uptime and downtime
* The `profile` operation prints information about the behavior of a single job across time
* The `top` operation prints information about CPU allocation

Finally there are some debugging operations:

* The `metadata` operation prints meta-information about the `sample` table
* The `version` operation prints machine-readable information about sonalyze and its configuration,
  useful mostly for testing.
* The `help` operation prints high-level usage information.

Run `sonalyze <operation> -h` to get help about options for the specific operation.

All the operations that produce structured output take a `--fmt` option to control the output; run
with `--fmt=help` to get help on formatting options and the available output fields.

Command line parsing allows for `--option=value`, `--option value`, and also to spell `-option` with
a single dash in either case.  It is however not possible to run single-letter options together with
values, in the manner of `-f1w`; it must be `-f 1w`.

### Data insertion operations

The `add` operation is used to ingest data into the database.  In addition to an option
`--data-path` that identifies the cluster directory, the main mode identifies the type of data.  The
data are self-identifying and are always read from stdin.

`--sample`

  The data are "free CSV" data coming from `sonar ps`, representing samples for one or more systems, zero
  or more records

`--sysinfo`

  The data are JSON data coming from `sonar sysinfo`, identifying one particular system, one
  record exactly.

`--slurm-sacct`

  The data are "free CSV" data coming from `sacctd`, which extracts data from the Slurm databases
  using `sacct`.

`--data-path=<path>`

  Root directory for log files, overrides the default.  The default is the `SONAR_ROOT` environment
  variable, or if that is not defined, `$HOME/sonar_logs`.

### Analysis and data extraction operations

#### Overall operation

The program operates by phases:

* reading any system configuration files
* computing a set of input log files
* reading these log files with record filters applied, resulting in a set of input records
* aggregating data across the input records
* filtering the aggregated data with the aggregation filters
* printing the aggregated data with the output filters

Input filtering options are shared between the operations.  Aggregation filtering and output options
are per-operation, as outlined directly below.

Sonalyze can also act as a client toward a server that runs sonalyze on its behalf against a data
store on a remote host, see the section "REMOTE ACCESS" further down.

#### Log file computation options

`--data-path=<path>`

  Root directory for log files, overrides the default.  The default is the `SONAR_ROOT` environment
  variable, or if that is not defined, `$HOME/sonar_logs`.

`-- <logfile> ...`

  If present, each `logfile` is used for input instead of anything reachable from the data path;
  the data path is ignored.

#### System configuration options

`--config-file=<path>`

  Read a JSON file holding system information keyed by hostname.  This file is required by options
  or print formats that make use of system-relative values (such as `rcpu`).  See the section
  "SYSTEM CONFIGURATION FILES" below.

#### Record filter options

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

`--exclude-system-jobs`

  Exclude jobs with PID < 1000.

`-j <job#>`, `--job=<job#>`, `--job=<job#>,<job#>,...`

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


#### Job filtering and aggregation options

These are only available with the `jobs` command.  All filters are optional.  Jobs must pass all
specified filters.

`--merge-all`, `--batch`

  Aggregate data across hosts (this would normally be appropriate for systems with a batch queue,
  such as Fox).  Normally this flag comes from the config file; this is an override.

`--merge-none`

  Never aggregate data across hosts (this would normally be appropriate for systems without a batch
  queue, such as the UiO ML nodes).  Normally this flag comes from the config file; this is an
  override.

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

#### Load filtering and aggregation options

These are only available with the `load` command.  All filters are optional.  Records must pass all
specified filters.

`--hourly`, `--half-hourly`

  Bucket the records (half)hourly and present averages (the default).  Contrast `--daily` and `--none`.

`--daily`, `--half-daily`

  Bucket the records (half)daily and present averages.  Contrast `--hourly` and `--none`.

`--none`

  Do not bucket the records.  Contrast `--hourly` and `--daily`.

`--group`

  Sum bucketed/averaged data by time step across all the selected hosts, yielding an aggregate for this
  group/subcluster of hosts.  Requires bucketing other than `--none`.

#### Sacct filtering and aggregation options

Since these are not sample records they have their own filtering rules.

The default is to print "regular" jobs, ie, not Array jobs or Het jobs.  Select the latter groups
with `-array` and `-het`.

`--state`

  Select jobs by termination state: `COMPLETED`, `CANCELLED`, `DEADLINE`, `FAILED`,
  `OUT_OF_MEMORY`, `TIMEOUT`

`--host`

  Select by node names.

`--user`

  Select by user names.

`--account`

  Select by account names.

`--partition`

  Select by partition name.

`--job`

  Select by job number.  If the job number is the overarching ID of an array or het job then all
  the subjobs are selected, otherwise only the one job.

`--all`

  Turn off some default filters.

`--min-runtime`, `--max-runtime`

  Filter by job elapsed time.

`--min-reserved-mem`, `--max-reserved-mem`

  Filter by amount of memory requested for the job.

`--min-reserved-cores`, `--max-reserved-cores`

  Filter by the number of cores requested for the job, this is usually node count times cores per node.

`--no-gpu`

  Select only jobs that used no GPU.

`--some-gpu`

  Select only jobs that used some GPU.

`--regular`

  Select only regular jobs (default).

`--array`

  Select only array jobs (including subjobs of an array job).

`--het`

  Select only het jobs (not implemented yet).

#### Job printing options

`--breakdown=<keywords>`

  NOT CURRENTLY IMPLEMENTED.

  For a job, also print a breakdown according to the `<keywords>`.  The keywords are `host` and
  `command` and can be present in either order.  Suppose jobs are aggregated across hosts (with
  `--merge-all`) and that the jobs may run as multiple processes with different names.  Adding
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


#### Load printing options

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

#### Uptime printing options

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

#### Profiling printing options

`--fmt=<format>`

  Format the output for `profile` according to `format`, which is a comma-separated list of keywords,
  see OUTPUT FORMAT below.

#### Top printing options

There are no options.  This verb is really a WIP.  The only output at the moment is a human-readable
representation of how busy the node was in a given interval, giving a visual indication of load
balancing.  CSV, awk and JSON need to be implemented but are not yet.

#### Sacct printing options

There are no options for printing beyond `-fmt`, but a lot of selection options.

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

## INI FILES

The file `$HOME/.sonalyze` can contain defaults for some arguments.  The file is a line-oriented
sequence of sections starting with a header and containing definitions, environment variables are
expanded.  Comment lines are allowed.  A typical file:

```
# Standard setup

[data-source]
remote=https://naic-monitor.uio.no
auth-file=$HOME/.ssh/sonalyzed-auth.netrc
```

Currently the only section is `[data-source]` and valid keys are `remote`, `auth-file`, `cluster`,
`data-dir`, `from`, and `to`.

In the future it seems likely that sections could be added to provide defaults for
e.g. `[record-filter]` and for individual verbs, e.g. `[jobs]`.

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
command and don't always have the same meaning for different commands (sorry).  In addition to the
field names there are control keywords that determine the output format:

* `fixed` forces human-readable output in fixed-width columns with a header
* `csv` forces CSV-format output, the default is fixed-column layout
* `csvnamed` forces CSV-format output with each field prefixed by `<fieldname>=`
* `json` forces JSON-format output (without any header, ever)
* `awk` forces space-separated output, with spaces replaced by `_` within fields
* `header` forces a header to be printed for CSV; this is the default for fixed-column output
   (and a no-op for JSON)
* `noheader` forces a header not to be printed, default for `csv`, `csvnamed`, and `json`
* `nodefaults` applies in some cases with some output forms (notably the `parse` command with
  csv or JSON or awk output) and causes the suppression of fields that have their default values)
* `tag:something` forces a field `tag` to be added for each record with the value `something`

Run with `--fmt=help` to get help on formatting syntax, field names, aliases, and controls, as
well as a description of the default fields printed, sort order, etc.

A "Unix timestamp" is the number of seconds since 197-01-01T00:00:00UTC.

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

### Performance

#### Filter by host/node when possible

Dealing with large data sets can be comparatively slow, especially if the set has to be queried many
times for a report - the amount of caching of intermediate data in the server is variable, for good
and bad reasons.

The `adhoc-reports/heavy-users.py` script is a case in point.  It performs a single `jobs` query to
extract a number of jobs and users and then performs a `profile` query per job to retrieve a profile
of the job, which is then further processed.  It pays to make the profile query as precise as
possible to winnow the initial set of records to be processed.

The database is organized by cluster, by date, and by host, and these are the keys used when
constructing the initial set of records.  Cluster and date range are mandatory filters for every
query, but if the query does not filter by host (node), then records for all nodes will be read.
All other filtering (user name, job number, command, ...) happens subsequently.

Even on Fox, which has a modest number of nodes, there may be a large number of records for an
interesting date range such as a month.  Filtering by the node name for a job, which should always
be possible when constructing a profile and in many other cases, can reduce the time for each query
by a factor of 20 or so.  On a larger system the factor will be correspondingly larger.

#### Batching queries

In principle, queries that use the same base set of records can be batched so that the base set is
constructed once, queries are run against it, and multiple results are returned.  (One might wish
that caching in the server would automatically do this, but it does not.)

No such batch option exists at this time.  Make some noise if this is starting to look like an
important feature.

## REMOTE ACCESS

Sometimes the log files are not stored locally but on a remote host to which we cannot log in.  If
the `sonalyze daemon` (aka `sonalyzed`) server is running on that host, we can run sonalyze locally
and ask it to contact that server and run sonalyze against the data stored there.  The server
responds to HTTP `GET` and `POST` requests; the local sonalyze constructs and sends the appropriate
request (currently via curl).

To do so, use the `--remote` argument to provide an http URL for the remote host and the `--cluster`
argument to name the cluster for which we want data:

```
$ sonalyze jobs --remote http://some.host.no:8087 \
                --cluster ml \
                --auth-file some-file.netrc \
                -f 20w \
                -u - \
                --some-gpu \
                --host ml8
```

The auth-file holds the identity information of yourself, keep this file secret.  In this case, the server
must have been told about this identity.  See the server manual for how to set that up.

The auth file can either have a `username:password` pair or (much better) be on the .netrc format.
It must have a single line of text.  The netrc format is `machine <machine> login <username>
password <password>`.  The advantage of this is that the file can be passed to curl, while a
username:password combination has to be passed on the command line and may be visible to other
users.  The username:password facility will probably be removed eventually.

In the case of remote access, the server supplies the `--data-path` and `--config-file` arguments
based on the `--cluster` argument, so the former must be omitted from the local command invocation.
Additionally, no trailing file arguments (`-- filename ...`) are allowed here.

For the `add` operation, the credential supplied with `--auth-file` will be matched against the
server's "upload credentials" database.
