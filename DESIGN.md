# Jobanalyzer: Design, implementation, and discussion

## Architectural overview

There is a variety of use cases (see [`README.md`](README.md)) and as a consequence, a fairly
elaborate architecture.  But in brief, the system is built around a system sampler and a pool of
samples, with a stack of increasingly user-friendly, stateful, and specialized tools to process
those samples.

We use [`sonar`](https://github.com/NordicHPC/sonar) as a general system sampler.  Sonar runs
periodically (typically every 5m or so, by means of `cron`) on all interesting hosts, and indeed on
every node in a cluster; it samples the hosts' state when it runs, and writes raw sample data to
files in a directory tree.  These samples need careful postprocessing and additional context to make
much sense.

Sonar is stateless and contextless: it samples the system state, tries to clean up what it can in
the absence of any context, and appends data to a log file.

We use [`sonarlog`](sonarlog) (in this repository) to ingest, contextualize, enrich, tidy up, and
filter the Sonar logs.  Sonarlog produces "sample streams", which are clean, self-consistent, and
chronologically sensible per-process or per-process-cluster sample data.  Sonarlog runs as a component
of `sonalyze`, see next.

We use [`sonalyze`](sonalyze) (in this repository) to aggregate, merge, query, and format the sample
streams to present meaningful summaries of "jobs".  See [`sonalyze/MANUAL.md`](sonalyze/MANUAL.md)
for instructions about how to run it, and [`README.md`](README.md) for some sample use cases.

Sonarlog and Sonalyze are stateless, but they have context: the entire sample stream (or the window
on those streams that has been selected) is available for inspection, and these components can and
do compute inter-record values.

Part of the complexity in the system up to this level stems from its generality: it works on both
single-node and multiple-node systems, on systems with or without a batch queue, and for jobs that
consist of a single command as well as those that consist of many different commands.

Anyway, there are many options available to Sonalyze to make it or Sonarlog select time windows,
sample records, jobs, and output formats.  See [`sonalyze/MANUAL.md`](sonalyze/MANUAL.md) or run
Sonalyze with `--help`.

Built on top of Sonalyze there are shell scripts that run periodically to run Sonalyze on the Sonar
logs and to produce further logs.  These scripts are system-specific; the ones for the UiO ML nodes
are in [`production/ml-nodes`](production/ml-nodes) and correspond in some cases directly to use
cases in [`README.md`](README.md).  For example, the [`cpuhog.sh`](production/ml-nodes/cpuhog.sh)
script runs a Sonalyze query that looks for longish-running programs that use a lot of CPU and no
GPU, and therefore technically abuse the ML nodes, which are meant for GPU-heavy computations.

Sonalyze being stateless, the scripts are also stateless.

Then there is a tool [`naicreport`](naicreport) that is a user-friendly superstructure for most of
the foregoing: it ingests the logs produced by analysis shell scripts, it runs Sonalyze directly,
and produces human-readable and machine-readable reports and data, for further incorporation in
emails, plots, and so on.

Unlike the other tools, Naicreport has state, allowing it to avoid sending reports it has sent
before, to keep track of when certain problems appeared and when they were last seen, and similar
things.

Then there are scripts built on top of Naicreport that run it periodically and upload its reports
(all JSON) to a web server.

The web server has simple presentation logic for the JSON data, and always works on whatever data
have been uploaded - it has no other state.


## Production setup

### ML Nodes

See [production/ml-nodes/README.md](production/ml-nodes/README.md) for instructions about how to set
up and run everything.  In brief, all scripts and programs live in a single directory with
subdirectories for data and reports, and cron jobs run scripts at sensible intervals to do all the
work.


## Implementation overview

Sonar runs often and on systems that are used for HPC, and needs to be lightweight.  Currently it
runs in about 50-100ms (not including the overhead of `cron`); further reductions are likely
possible.

The other tools can run on any system that has access to up-to-date Sonar output, and they don't
have to be very fast if they aren't being run on the HPC nodes.

Sonar produces output in "free CSV" form, that is, using CSV syntax but with each field named by a
`fieldname=` prefix and a potentially variable number of fields per record, allowing fields to be
added and removed over time and several versions of Sonar to be in use at any time.

Sonar is written in Rust (a sensible choice, and one made some time ago).  Sonarlog is also written
in Rust, specifically so that it can be shared between Sonalyze and another tool,
[`jobgraph`](https://github.com/NordicHPC/jobgraph).  Jobgraph and Sonalyze are also written in Rust
and Sonarlog can be used as a component of these tools, it is not run standalone.

Sonalyze can produce both human-readable and free CSV output, as different use cases call for
different formats.

Naicreport is written in Go for greater flexibility (the rigidity of Rust's ownership system and
manual memory management get in the way of getting things done).  Its state files are in free CSV
form.  It can produce human-readable or JSON output.

Logic is pushed into Naicreport and Sonalyze when it is sensible and possible; the surrounding shell
scripts are kept very simple.


## Various considerations

### Privacy 

Currently the log files are publically accessible on any system that mounts the disk where the logs
are stored, and this is by design: all the "user" use cases require this.  However, there are
concerns around GDPR/privcy as well as security.  The log contains a history of runs, keyed by UID
and time, and part of the command line for a process.  Thus the user's activities may be tracked and
exposed without consent, and should there be a secret embedded in the command name it may be
exposed.

On the one hand, this information is not privileged to other users of the system: anyone running
`top` or `ps` would see the information.

On the other hand, information in the log may become viewable from outside the system - if the disk
is mounted elsewhere, or as part of job summaries uploaded to publically visible servers.

### What are "requested resources"?

Several use cases above compare the consumed resources to the (explicitly or implicitly) requested
resources, or to the available resources.  Thus, on systems where it makes sense the log (or an
accompanying log) must also contain the requested resources.  For example,

* On ML nodes with expensive GPUs, the GPUs are implicitly requested.
* For scalability analyses, if a program can't make use of the machine it's running on (the
  "implicitly requested resources") then it's not going to help moving it to a larger system.

At the moment, the "requested resources" for the ML nodes are encoded in the script that produces
the reports about resource usage, [`cpuhog.sh`](production/ml-nodes/cpuhog.sh).

What does it mean for a job to be using a "lot" of CPU and a "little" GPU?

Consider a machine like ml6 which appears to have 32 (hyperthreaded) CPU cores, 256GB of RAM, and
eight RTX 2080 Ti cards each with 10GB VRAM.

Which of these scenarios do we care about?

* A job runs flat-out on a single CPU for a week, it uses 4GB RAM and no GPU. (We prefer it to move
  to light-HPC/Fox but we don't care very much, *unless* there are many of these, possibly from many
  users.)

* A job runs flat-out on 16 cores for a week, it uses 32GB of RAM and no GPU. (We really want this
  to move to light-HPC/Fox.)

* Like the one-CPU case, but it also uses one GPU for most of that time.  (I have no idea.)

* Like the 16-CPU case, but it also uses one GPU for most of that time.  (I have no idea.)

* Like the 16-CPU case, but it also uses several GPUs for most of that time.  (It stays on ML6,
  unless it's using a lot of doubles on the GPUs, in which case it should maybe move to ML8 with the
  A100s?)

It is likely that there needs to be a human in the loop: the system generates an alert and the human
(admin) can act on it or not by alerting the user.  I guess in principle this is an interesting
machine learning problem.

### Other tools

* There are good profilers already, but generally you need to commision the profile when the job
  starts, and sometimes you must rebuild the code for profiling.  Traditional profilers do not speak
  to most of the use cases.

* Some job monitors may do part of the job, for example, `nvtop` will show GPU load and process ID
  and gives a pretty clear picture of whether the job is running on the GPU.  (Like `htop` and `top`
  for the CPU.)  These monitors can be started after the job is started.  In fact, `nvtop` shows
  both GPU and CPU load and a quick peek at `nvtop` is often enough to find out whether a busy
  system is being used well.

* `nvtop` also works on AMD cards, though only for kernel 5.14 and newer.  (There is also
  https://github.com/clbr/radeontop which I have yet to look at.)

* `cat /proc/loadavg` gives a pretty good indication of the busyness of the CPUs over the last 15
  minutes.

* `nvidia-smi` can do logging and is possibly part of the solution to generating the log.  See
  SKETCHES.md for more.

* `rocm-smi` may have some similar capabilities for the AMD cards.

* The `jobgraph` tool, augmented with a notion of what a "job" means on the ML and light-HPC
  systems, can be used to address the three "Development and debugging" use cases: it can take a job
  (or a set of jobs, with a little work) and display their resource consumption, which is what we
  want.

* The code that creates the load dashboard on ML nodes is [here](https://github.uio.no/ML/dashboard).

* Sigma2 uses RRD for some things but this is (apparently) more a database manager and display tool
  than anything else.

* We have something running called Zabbix that is used to monitor health and performance but I don't
  know how this works or what it does.

* Open XDMod seems like a comprehensive tool but may be dependent on having a job queue.


