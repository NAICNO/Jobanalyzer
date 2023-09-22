# Cross system Jobanalyzer

Jobanalyzer: Easy-to-use resource usage reporting and analyses.

## Overview

Jobanalyzer is a set of tools providing the following types of services:

- for admins: monitoring of current and historical utilization, as well as usage patterns
- for users: first-level analyses of computation patterns, with a view to appropriate system 
  use and scalability - cpu use, gpu use, communication use

The tool set is expected to grow over time.

Current tools are based on a system sampler, and provide information based on collected samples.
See [DESIGN.md](DESIGN.md) for more information on the technical architecture and its
implementation.


### Admins

Admins will come to Jobanalyzer via [its web interface](https://axis-of-eval.org/naic).  The current
interface is bare-bones and consists only of a node-centric load dashboard, allowing the current and
historical load of each node to be examined.  Only the UiO ML nodes are currently represented.

Data for the web interface are produced firstly by periodic analysis by the low-level `sonalyze`
tool and secondly by the higher-level `naicreport` tool, and the results of these analyses are
uploaded periodically to the server.

The web interface will be extended with dashboards for alerts for jobs that are using the system
inappropriately ("cpu hogs") and jobs that are terminated but hanging around anyway and are holding
onto system resources ("dead weight").  (Currently those alerts are emailed.)


### Users

Users will currently come to Jobanalyzer via its command line interface (there is room here for a
web interface or other GUI but that's down the road).  The primary interface is via the low-level
`sonalyze` tool.  This tool can be hard to use effectively, but does serve many use cases as
described below.

In practice, it is likely that common use cases for users should be packaged up and provided
directly via the higher-level `naicreport` tool.


## Sample use cases

The use cases span a usage spectrum from "pure production" to "partly development" to "systems
administration".  In principle, the hardware spans the spectrum: personal systems, ML nodes, UiO
light-HPC, Fox, Colossus, national systems. (Unclear: LUMI.)  The usage spectrum is large enough
that this may be multiple tools, not a single tool.

The section headings below are the names for these use cases referenced elsewhere, including in
code when appropriate.

Below, there are examples of how `sonalyze` is used to address the use cases by querying the system
sample data.  Remember that there are higher-level interfaces than this.


### Admin use cases

#### `cpuhog`

This is an automatic monitoring and offloading use case.

> User X runs a job on an ML node but the job does not use the GPUs, yet X's CPU usage is such that
> other users who could use the GPUs do not use the machine because the CPUs are overloaded.  X should
> move to a non-GPU system such as Fox or the GPU-less light-HPC systems, but user X is unaware of the
> problem.  X or admins should be alerted to the problem so that X can be made to move.

There are some issues with the problem definition; what is "CPU usage such that ... the CPUs are
overloaded"?  Clearly if the user uses, say, half the CPUs but is alone on the system, there may not
be an actual problem.

For the sake of simplicity, let's say that using 10% of the CPUs or more at peak without using any GPU
violates this policy.  (This is a little primitive but good enough for an experiment.)  Then this is
expressed as a query against the sonar logs:

```
sonalyze jobs --config-file=ml-nodes.json -u- --no-gpu --min-rcpu-peak=10 --min-runtime=10m --fmt=tag:cpuhog,...
```

(Of particular note here is that `-u-` selects all users; for the `jobs` command, the default in
most cases is the user running the command.)

Suppose there is a job with ID 12345 that triggers this query.  To examine the job's behavior in
depth, one can currently run `sonalyze load`:

```
sonalyze load --job=12345
```

which will show hourly data for the job over the last 24h (or add `--from 2d` for the last 48h, etc).

#### `deadweight`

This is an automatic or manual monitoring use case.

>Zombie jobs and other leaks hold onto GPU or main memory, or use GPU or CPU resources.  Systems
>administrators should be alerted to this fact, or should be able to use a tool to quickly discover
>these situations.

The support for this use case is a little thin but the following command currently does part of the
job:

```
sonalyze jobs --zombie
```

#### `current_utilization`

This is a manual monitoring use case.

>Admin Y wants to view the current load of a shared server.

Here a question is whether the admin cares about total load or just the load from long-running jobs.
Probably it's the latter since the former could be had with `htop` or similar tools.  The `load`
command shows system load; the `--last` switch shows the last sample for each host only:

```
sonalyze load --last
```

Or filter by host name:

```
sonalyze load --last --host=ml[6-8]
```

#### `historical_utilization`

This is a manual monitoring use case.

>Admin Y wants to view historical utilization data of a shared server.

Here a question is whether the admin cares about total load or just the load from long-running jobs.
At present, `sonar` and `sonalyze` only offer the latter (to do the former we have to log rusage data
in sonar, not hard to do).

Here's the daily average CPU and GPU utilization for the last year.  (Hourly averages may be more
meaningful but would create too much data for the year.)

```
sonalyze load --from=1y --daily --fmt=header,cpu,gpu
```
Note these are "absolute" values in the sense that, though they are percentages, the reference for
100% is one CPU core or GPU card.  If you instead want values relative to the system, you need to
ask for that, and you need to provide the system configuration, here are hourly system-relative
averages for the last three days:

```
sonalyze load --from=3d --fmt=rcpu,rgpu --config-file=ml-systems.json
```

### User use cases

### `verify_gpu_use`

This is a development and debugging use case.

>User X runs an analysis using Pytorch. X expects the code to use GPUs. X wants to check that the
>code did indeed use the GPU during the last 10 analyses that ran to completion.

In principle this is straightforward (if the jobs all ran within the last 24h):

```
sonalyze jobs -n 10 --completed
```

The default output has fields for GPU usage and can be easily inspected.

The bit about Pytorch is a little tricky though.  Currently we log the name of the executable being
run, so the above could be filtered by, say, `python` (using `--command=python`).  Mostly that's not
very useful (it's all Python).  We do not log the entire command line (for both privacy and
technical reasons), nor do we log files accessed by the job (this would require a very different
level of logging amounting to running the job under `strace` or similar.)

In practice, filtering by Pytorch will not be necessary.  It is possible to filter by minimum
runtime, or to show only jobs that used no GPU:

```
sonalyze jobs -n 10 --completed --min-runtime=10m --no-gpu
```

### `verify_resource_use`

This is a development and debugging use case.

>User X submits an HPC job expecting to use 16 cores and 8GB memory per CPU. Admins complain that X
>is wasting resources (the program runs on one core and uses 4GB). In order to debug the problem, X
>want to check how much resources the job just finished used.

```
sonalyze jobs -n 1 --completed
```

This will show the CPU and GPU utilization (in % of one core and % of one card), and memory use for
both, for example:

```
jobm     user      duration  cpu-avg  cpu-peak  mem-avg  mem-peak  gpu-avg  gpu-peak  gpumem-avg  gpumem-peak  host  cmd
1392113  username  0d 3h 0m  1199     1572      97       99        71       90        6           7            ml7   python
```

This says that it used 1199% CPU (ie about 12 cores' worth) on average and 1572% at peak; 97GiB RAM
on average and 99GiB at peak; 71% GPU (ie 2/3 of one card's worth) on average and 90% at peak; and
6GiB GPU RAM on average and 7GiB at peak.

(The `--completed` switch can be omitted usually, and sometimes it's in the way because the logs
have an imperfect notion of whether a job is still running or not.  It is only when a job stops
appearing in newer log records that `sonalyze` can conclude that the job has completed.)

### `verify_scalability`

This is a development and debugging use case.

>User X wants to understand a (say) matrix multiplication program written in C++ with an eye to
>whether it will scale to larger systems.

This has to be approached somewhat indirectly, but consider the example under `verify_resource_use`
above.  We don't yet have a notion of communication volume, but suppose this is not an issue and we
just want to know if the program will run on a larger multiprocessor.

To determine that, we should look at utilization relative to the machine's capabilities.  If the
user knows the capabilities (for example, she knows that ML7 has 32 hyperthreaded cores and 8 GPU
cards) it's plain that the job won't scale to a larger system, because it used only 12 cores and
less than 1 card on average.

More generally, the user can provide a configuration file to `sonalyze` that describes the machines
and can print (and query) on machine-relative data:

```
$ sonalyze jobs -n 1 --config-file=ml-nodes.json --fmt=job,user,duration,rcpu,rmem,rgpu,rgpumem
```

This yields:

```
job      user      duration  rcpu-avg  rcpu-peak  rmem-avg  rmem-peak  rgpu-avg  rgpu-peak  rgpumem-avg  rgpumem-peak
1392113  username  0d 3h 0m  19        25         38        39         9         12         1            1
```

and it's now fairly obvious that the system is not maxed out.

Other tools (`perf` and so on) should then be brought to bear on the root causes for why the system
is not maxed out.

### `thin_pipe`

This is an automatic or manual monitoring use case.

>User X runs a job on several nodes of a supercomputer and the jobs communicate heavily, but the
>communication does not use the best conduit available (say, uses Ethernet and not InfiniBand).  X
>or admins should be alerted to the problem so that X can change the code to use a better conduit.

There is currently no support for this (no logging of communication bandwidth in sonar).


## Non-use cases (probably)

* User X is developing new code and sitting at the terminal and wants to view GPU, CPU, and memory
  usage for the application, which is running.  For this X can already use `nvtop`, `nvitop`,
  `htop`, and similar applications.

* Admin Y is wondering what the current total load is on the system.  For this Y can use `nvtop`,
  `nvitop`, `htop`, and similar applications.

* In general, traditional "profiling" use cases during development (finding hotspots in code, etc)
  are out of bounds for this project.

