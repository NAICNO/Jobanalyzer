# Requirements

This is based on discussion in [ticket 92](https://github.com/NAICNO/Jobanalyzer/issues/92), older
sketches and plans, and on more recent discussions in the development org about how to turn our
prototypes into a unified, useful system.

This document exists primarily to create agreement within the development org.

## Overall view

Jobanalyzer is a set of tools providing the following types of services for the following user groups:

- For systems admininstrators: monitoring of current and historical utilization, as well as usage
  patterns of the systems and installed software; guidance in moving users' computations from one
  system to another to improve performance or even out load spikes.

- For users and support staff: first-level analyses of computation patterns, with a view to
  appropriate system use and scalability - cpu use, gpu use, i/o; guidance in moving a computation
  from one system to another to improve performance and make appropriate and optimal use of
  resources for a task.  Note there are several subgroups each of users and support staff.

- For the public, decision makers, and public agencies: utilization statistics and other high-level
  data about the value of the system, this could be pretty basic (X% utilization during the last
  quarter) or sophisticated (system used for N months in project X which published papers P and Q).

A list of detailed use cases and concrete requirements is presented later; the following subsections
provide a high-level breakdown.  There is a list of non-cases at the end.

### Systems administrators

Administrators will partly come to Jobanalyzer via [its web interface](https://naic-monitor.uio.no).
The web page presents a management console with data about each of the systems under its control.
These data will typically include:

- Current and historical system load (computed in a job-centric way)

- Current and historical jobs

- Current and historical uptime

- Alerts for users and/or jobs that should not be on the system because they are not using the
  resouces they requested, a typical case is that a job running on a machine meant for
  GPU/accelerator computation is not using the accelerators

- Detailed job data including profiles for individual jobs

The data on this console will mostly come from periodic reports, that is, a query or dashboard will
display the contents of a report already generated.  Some reports will update as often as hourly,
others more rarely.

Adminstrators will partly also use a command line interface, that allows for a more in-depth
analysis of users' jobs with an eye toward moving jobs and users from the system they are on to a
machine that better fit their jobs.

In this context, the "load" on the system is the sum of the load of the users' jobs, not necessarily
the true load, which would also include the operating system and similar "overhead".

(Probably systems administrators are not going to be the primary users of Jobanalyzer.  They will
have their own monitoring and administration tools and will want to have a system-centric view of
system load.  Alerts for users or jobs that misuse the system will likely go to support staff, not
to sysadmins.)

### Users and support staff

Users and support staff will currently come to Jobanalyzer via its command line interface (there is
room for a web interface or other GUI in the future).  The user will query the system for
information about her own running or completed jobs to examine them for system utilization and
scalability; the support staff, working on a case, will examine the submitting user's jobs.

There may be an interface that allows a job to be compared post-mortem with a database of benchmark
results for the purpose of finding a machine that is a good fit for the next run of the job.

We can identify subgroups of users:

* Developer types: Scientists and software types that build programs, ML models, and similar
* "Hands-on" end-user types: Users of finished models, builders of pipelines, scientists that
  use tools such as R, Jupyter, MATLAB to explore their problem space
* "Production" users: People who run production jobs of otherwise finished software (that they
  may have written themselves)
* "External" users: SMB users that develop and run their work on national systems but may have
  a different notion of service, uptime, timeliness of answers, and cost than scientist types
* "Expert" users: a cross-cutting category, these are familiar with advanced SW development tools
  and the ins and outs of tuning for a system, but may come to Jobanalyzer for a first assessment
  of a problem and post-hoc data

Similarly there are several levels of support staff:

* Front-line support staff may be similar to developer types: they know some things about
  how the systems work and can benefit from high-level post-hoc analyses of jobs
* More advanced support staff are similar to both expert users and sysadmins and can
  benefit from analyses of workloads, users, and systems to aid them in helping users
  place and run their jobs.

### The public and decision makers

The public and decision makers will come to Jobanalyzer mostly through prepared reports but there
could be a high-level web interface too, a type of "load dashboard light".  The reports of interest
will include:

- Current and historical system uptime and utilization (including, say, utilization of accelerators)

- High-level view of the types of jobs run, if we can manage it.  Obvious breakdowns are by discipline,
  by task (training vs inference, say), by framework; but likely there are more dimensions

- System demand, ie, number of users requesting access that are denied access; queue lengths if
  appropriate; and so on - again on an aggregate level

Reports will usually be fairly long-term: monthly, quarterly, annually, or longer.

### Security and privacy

The final system must not reveal personally identifiable information to an unauthorized party.  For
example:

- A user should only be able to examine her own jobs in any detail.

- Any publicly available load/alert dashboard should redact user names, project numbers,
  affiliations, and possibly also command names; such privileged information should only be
  available to logged-in systems administrators

- Possibly a number of dashboards should simply not be available to the public.

In practice, raw log data must not be accessible to users; log data must be processed and laundered
on a protected server and not on a client; and there must be a clear sense of which data require
which level of privilege and the privileges must be enforced on the protected server.

### Demands placed on NAIC systems

All systems under NAIC's purview should run Jobanalyzer's data gathering component so that jobs and
system load can be analyzed by NAIC personnel.

These systems vary along multiple dimensions: in their size (number of nodes), their power (number
of cores and amount of memory), their technology (cpu architecture, the presence of accelerators,
and accelerator architecture), their job management (batch queues vs free-for-all), and their
ownership and access (ranging from single-university access to LUMI).

If it is not possible to run Jobanalyzer continually in the background on a system (for example, if
NAIC does not control the system, but does make use of it) then it should be possible to run the
data gathering in parallel with the job, while the job is running, so that the job can be analyzed
in the standard manner during its run and after its completion.

## Artifacts produced

THIS IS BOTH EVOLVING AND STALE.

At least these:

- web-based load dashboard that shows current or near-current and historical loads of all known
  systems, specific UI TBD (see Use cases, below)

- web-based alert dashboard that shows alerts for various conditions that are of interest, such as
  inappropriate system use; specific triggers are TBD (see Use cases, below), with functionality for
  extracting reports of users or jobs that regularly cause alerts

- command-line based tool to examine specific systems and jobs in detail, breaking down jobs
  into per-system components and systems into clusters of jobs running on them

- command-line based tool to produce reports on software used over a period of time, output format
  TBD

## Policy

We are going to need to become precise when we talk about "a lot" of resources and "many" users (or
indeed "few" resources and "few" users).  The values of these terms depend on the system in question
and maybe even on the time of the year (bigger resource crunch during the term than outside it).

For this the monitoring tools need inputs from SLAs and RFK (research allocation committee).  What
this means is that when users ask for resources under NAIC, they will indicate what resources
(e.g. GPU-hours) they need.  Their usage should somewhat align with what they asked for.  Not all
systems work this way: the current ML nodes do not have resource allocation, it's a free-for-all.

TODO: More work needed.

## Scope

Jobanalyzer should not try to do the job of existing good tools.  For example, `htop` and `nvtop`
will be better at displaying moment-to-moment load.  A profiler such as `perf` will do better at
finding bottlenecks in the code.  There are also many other, much better tools than those.

What we can hope to do is provide high-level, easy-to-use tools that will collect data across time,
enough to give a view of the systems' utilization and, help with the initial diagnosis of
scalability issues, help with systems management and load balancing, and report on the systems
health and use.

## User stories that will be supported

THIS IS NOT UP-TO-DATE.  In particular it fails to pin stories properly to user groups and it only
coarsely distinguishes support staff from admins and users.  Also not all stories from recent
discussions may be represented.

User stories are given identifier-like names so that they can more easily be referenced elsewhere.
*A* is a systems administrator.  *U* is a user.  *P* is a member of the public.  *D* is a decision
maker.  NAIC management is lumped in with the administrators, mostly.

Unless otherwise stated, we are assuming that there is continual data collection (sampling) on all
systems.  For some use cases, more is needed.

I have annotated these with the Work Packages they most likely pertain to, see [Gitlab
issues](https://gitlab.sigma2.no/naic/wp2/identify-most-resource-intensive-users/-/issues).

### "Admin" user stories

#### `adm_systems_at_a_glance`

*Story 1 (WP2.3.5):* *A* wants a quick overview of the state of the various systems: whether they
are up, and how busy they are.

To retrieve these data, *A* goes to the web-based overview dashboard and selects the system(s) and
time window, and a number of plots will be shown, one per system, all on one page.  For multi-node
systems, all nodes are selected and aggregate data are shown.  The quantities plotted are
system-relative cpu, main memory, gpu, and gpu memory usage.  (TBD: queue length) If a system is
down during a period this is marked clearly and distinguished from zero load.

*A* can select time periods from the set last 6 hours (moment-by-moment data), last day (hourly data),
last week (hourly data), or last month (daily data).  Moment-by-moment data are as fresh as possible,
hourly data are at most an hour out of date, and daily data at most a day out of date.

*Story 2 (WP2.3.5):* *A* wants an overview of processes that are stuck (zombie, defunct) or holding
onto GPU resources (GPU lists them as active but they are dead).

To retrieve these data, *A* goes to the web-based overview dashboard, selects system and time window
(usually only the last few days are interesting), and gets a textual report of jobs that appear to
be dead weight and the reason for this.

*NOTE*, Story 2 is probably irrelevant now, this is the domain of traditional systems management
software, which will do this job better.

#### `adm_historical_system_load`

*Story 1 (WP2.3.5):* *A* wants to know the historical load on a system (single node or all nodes) to
see how busy it has been.  *A* is interested in short-term historical utilization, on the order of
hours to weeks, in a time window ending at the present time.  *A* is interested in the load as a
percentage of the system resources: cpu, main memory, gpu, and gpu memory.  (TBD: queue length)

To retrieve the load, *A* goes to the web-based load dashboard and selects the appropriate system and
time period, and a plot is produced showing the interesting quantities.

The quantities need not be 100% up to date - data that are a couple of hours old are acceptable.

*A* can select time periods from at least the set last day, last week, or last month.

On multi-node systems, *A* can select at least between one specific node or all nodes.

On systems with many nodes, or when there are many systems to select between, *A* can type in the
system name or part of it to quickly select the system.

*Story 2 (WP2.3.5):* *A* finds something interesting in the plot and wants to communicate this to a
coworker or archive the situation for later.  *A* needs an artifact that can be used later.

To get this artifact, *A* can (in the worst case) take a screenshot of the plot.  That requires the
screenshot to contain the information about the time range and system selected.

Alternatively, *A* can ask for an image representing the situation (a little cleaner but basically
just the same as the screenshot), or *A* can copy the URL of the plot, provided this URL contains all
the necessary information and is not restricted to *A*'s use.

#### `adm_historical_uptime`

*Story 1 (WP2.3.5/2.3.6):* *A* wants to know the historical uptime of a system (single node or all
nodes).  *A* is interested in longer-term historical uptime, going back a few months or a year, up
to the present time.  *A* is interested in reporting the uptime to somebody.  Uptime is defined as a
time during which a user could have run a job.  A system can thus be running without being "up" per
se.

To retrieve the uptime data, *A* goes to the web-based uptime dashboard and selects the appropriate
system and time period, and a report is produced in a text area showing the interesting quantities.
*A* can copy the text into email or into a report.

The quantities need not be 100% up to date - data that are up to a day old are acceptable.

*A* can select time periods from at least the set last week, last month, last quarter, and last year.

On multi-node systems, *A* can select at least between one specific node or all nodes.

On systems with many nodes, or when there are many systems to select between, *A* can type in the
system name or part of it to quickly select the system.

### "Support" user stories

#### `sup_unused_capacity`

*Story 1 (WP2.3.3):* A course is held during a time period, placing a great deal of strain on Fox (say).
Meanwhile, Saga (say) has a lot of spare capacity.  *A* is interested in learning about this imbalance
so that he can tell instructors to move work to Saga.

*A* will primarily see this situation on the web-based alert dashboard, where a report that runs
(say) once a day detects this situation and sets a flag.  Once detected, the flag persists until the
situation has not been detected for some days. *A* will have to verify that the situation is still
the case before alerting instructors.

*Story 2 (WP2.3.3):* *A* does not want to have to poll the dashboard every morning, and wants to be
able to register to receive an alert by email when the situation arises.

*A* can state this preference in his profile on the console.

#### `sup_misfits`

This is a collection of very similar use cases pertaining to a job (or user) being a poor fit for
the system the job is running on.  In all cases, there will be policy definition questions (see the
section on Policy further up), and there are issues with determining that a user is preventing other
users from getting work done on systems where there is no queueing system.

In all cases, *A* will primarily see the described situation on the web-based alert dashboard, where
a report that runs (say) once a day detects that *U* or her job is a poor fit for the system and
adds *U* to a list of offenders with a reason for the offense.  The listing persists until the
situation has resolved.  It's OK if *A* has to verify the situation manually after the alert.

*A* will have to manually notify *U* about the issue.

In all cases, *A* does not want to have to poll the dashboard every morning, and wants to be able to
register to receive an alert by email when the situation arises.  *A* can state this preference in
his profile on the console.

For some of these use cases, near real time data / alerts is probably desirable - TBD.

*Story 1 (gpuhog) (WP2.3.1/2.3.2):* *U* is occupying and using a lot of GPU resources for a long
time (on a system without a batch queue).  This creates a situation where the work of one user
prevents many other users from getting work done.  *A* wants to move *U* to a system that has more
GPUs and a batch system.

*Story 2 (whale) (WP2.3.1/2.3.2):* This is a generalization of *Story 1* but applies to all systems:
*U* is too big for the system - uses too much cpu, gpu, or memory, or for too long a time - and in
the way of other users.  And it may not be a matter of a single job that *U* is running, but the
aggregate of those jobs.  *A* wants to know this so that *U* can be moved to a larger system.

*Story 3 (cpuhog) (WP2.3.3):* *U* is occupying and using a lot of CPU resources for a long time but
not using any GPU, on a system reserved for GPU computation.  This creates a situation where the
work of one user prevents many other users from getting work done.  *A* wants to move *U* to a
system without GPUs.

*Story 4 (greedy) (WP2.3.3):* This is a generalization of *Story 3* but pertains more to systems
with batch queues: *U* has asked for resources but are not using them.  *A* wants to alert *U* to
this fact.

*Story 5 (mouse) (WP2.3.3):* An important wrinkle on Story 4 for systems with implicit resource
allocation (eg the ML nodes) is when a program is in the way because it is *too small* for a system.
ML8 is case in point: Everyone wants to use the fancy A100s.  But when a program uses a measly 8GB
VRAM, it could likely run on any of the ML nodes, yet when it runs on ML8 it is in the way of
programs that could use the 40GB provided by those cards.  (And when some of the other ML nodes
stand unused, this is Story 1 of `sup_unused_capacity` as well.)

*Story 6 (vampire) (WP2.3.1/2.3.2/2.3.3):* Another variation of *Story 3*, this problem occurs when
*U* does not run one big job that uses a lot of CPU and no GPU, but many smaller jobs, sometimes
overlapping, that together have the effect of being a big job in violation of the cpuhog policy.
See issue 55 for an exploration.  *A* wants to move *U* to a system without GPUs.

#### `sup_software_use`

*Story 1 (WP2.3.2):* *A* wants to know which software is being used by the users.  The reason may be
that the system is being upgraded and *A* doesn't want to upgrade dead software.

*A* should be able to at least run an ad-hoc report generator with a time window argument, it would
compile a textual report of the names of modules or programs that have been used.

*Story 2 (WP2.3.2):* *A* wants to know which software uses the most resources on a system.  The
reason may be to see if there is a way of reducing overall system load by upgrading or replacing
software.

*A* should be able to run another ad-hoc report generator with a time window argument, it should
compile a textual report of the names of programs that have been used.

(See issue #20 for more.)

#### `sup_users`

*Story 1 (WP2.3.1):* *A* wants to know which users and projects are using the system most heavily.
The reason might be to use them as a user group, or offer them special training, or other things.

#### `sup_management`

*Story 1 (WP2.3.6):* NAIC staff generates a report of system usage over time and sends to Management.

We want this to be high-level summary, probably pertinent to last month but maybe also including
historical trend for last quarter, half year maybe.

The report would contain information about users, projects, systems, software.


### "User" user stories

The "user" user stories are all very similar - the user runs a job and needs to find out what the
utilization was.  The difference between the stories is the focus the user has: whether it's just
for information, to examine hardware usage, to resolve performance problems, to examine scalability,
or to assess the appropriateness of running the program on the system in question.  As such these
user stories are also very close to the `sup_misfits` stories, but coming from the user perspective,
not that of the admin.

(User stories overlap in large part with support staff stories, and we don't currently distinguish
between them.)

#### `usr_resource_use`

*Story 1 (WP2.3.5):* *U* submits an HPC job and wants to assess how the job used the available
hardware, without having any particular focus on anything in particular.  This is frequently the
first thing one does after porting a code to a new machine.

*U* breaks out the Jobanalyzer command line tool and asks to see the statistics on her last job that
completed.  There is probably a predefined, named query that will apply the most appropriate
options.  (Sabry brings up the similarity to the Slurm `seff` command, see the "Other tools" section
of [DESIGN.md](DESIGN.md).)

*Story 2 (WP2.3.5):* *U* submits an HPC job expecting to use 16 cores and 8GB memory per CPU. Admins
complain that *U* is wasting resources (the program runs on one core and uses 4GB). In order to
debug the problem, *U* wants to check which resources the job just finished used.

*U* breaks out the Jobanalyzer command line tool and asks to see the statistics on her last job that
completed.

*Story 3 (WP2.3.5):* *U* runs an analysis using Pytorch. *U* expects the code to use GPUs. *U* wants
to check that the code did indeed use the GPU during the last 10 analyses that ran to completion.

*U* again breaks out the command line tool and asks to see the statistics on her last 10 jobs that
completed.

*Story 4 (WP2.3.5):* *U* runs a complex pipeline on the ML nodes and observes that it crashes due to
apparent resource exhaustion.  *U* wants to find out how the individual commands behaved over time.

*U* can run the command line tool and ask to see the per-time-step statistics for the failing job.
If the detail is insufficient, *U* can first sonar in high-frequency mode in the background while
the pipeline is running to obtain more detailed data.

(For more detail on some of the cases, see the `verify_gpu_use` and `verify_resource_use` use cases
in README.md)

#### `usr_scalability`

*Story 1 (WP2.3.5):* *U* wants to understand a (say) matrix multiplication program written in C++
with an eye to whether it will scale to larger systems.

*U* breaks out the command line tool and looks at the statistics for a run of the program, relative
to the system configuration.  If the program is not using the system effectively then it probably
will not scale; if it does use all of the system, it might scale.

*Story 2 (WP2.3.5):* *U* runs a job on several nodes of a supercomputer and the jobs communicate
heavily, but the communication does not use the best conduit available (say, uses Ethernet and not
InfiniBand).  *U* should be alerted to the problem so that *U* can change the code to use a better
conduit.

TODO: To be developed.

*Story 3 (WP2.3.5):* Same thing, but for disk I/O.

TODO: To be developed.

(For more detail, see the `verify_scalability` and `thin_pipe` use cases in README.md, as well as
tickets #18 and #58.)

### "Public and decision-maker" user stories

#### `pub_periodic_reports`

*Story (WP2.3.6):* *D* wants quarterly and annual reports of how the systems are being used, in
order to decide whether to provide more money.

*D* asks *A* for this report... and maybe it turns into an admin use case.  Obvious things to want to
report on is system load, uptime, wait times, projects that were run, projects that were denied(!).

TODO: To be developed.


## User stories and use cases that will *NOT* be supported

By and large, these are use cases that are better served by other tools.

* User X is developing new code and sitting at the terminal and wants to view GPU, CPU, and memory
  usage for the application, which is running.  For this X can already use `nvtop`, `nvitop`,
  `htop`, and similar applications.

* Admin Y is wondering what the current total load is on the system.  For this Y can use `nvtop`,
  `nvitop`, `htop`, and similar applications.

* In general, traditional "profiling" use cases during development (finding hotspots in code, etc)
  are out of bounds for this project.

* In general, "systems monitoring" use cases - looking for nodes or clusters or interfaces that are
  down, job systems that don't work, etc - are out of bounds (even though some of the data could be
  used to build some such services).  Use traditional systems management software for this.
