# Requirements

(Based on discussion in [ticket 92](https://github.com/NAICNO/Jobanalyzer/issues/92) and also older
sketches and plans.)

The audience for this document is "everyone".

## Overall view

Jobanalyzer is a set of tools providing the following types of services for the following user groups:

- For systems admininstrators: monitoring of current and historical utilization, as well as usage
  patterns of the systems and installed software; guidance in moving users' computations from one
  system to another to improve performance or even out load spikes

- For users: first-level analyses of computation patterns, with a view to appropriate system use and
  scalability - cpu use, gpu use, i/o; guidance in moving a computation from one system to another
  to improve performance and make appropriate and optimal use of resources for a task

- For the public and decision makers: utilization statistics and other high-level data about the
  value of the system, this could be pretty basic (X% utilization during the last quarter) or
  sophisticated (system used for N months in project X which published papers P and Q)

A list of detailed use cases and concrete requirements is presented later; the following subsections
provide a high-level breakdown.

### Systems administrators

Administrators will partly come to Jobanalyzer via [its web interface](http://158.39.48.160/).  The
web page presents a management console with data about each of the systems under its control.  These
data will typically include:

- Current and historical system load

- Current and historical jobs

- Current and historical uptime

- Alerts for users and/or jobs that should not be on the system because they are not using the
  resouces they requested, a typical case is that a job running on a machine meant for
  GPU/accelerator computation is not using the accelerators

The data on this console will mostly come from periodic reports, that is, a query or dashboard will
display the contents of a report already generated.  Some reports will update as often as hourly,
others more rarely.

Adminstrators will partly also use a command line interface, that allows for a more in-depth
analysis of users' jobs with an eye toward moving jobs and users from the system they are on to a
machine that better fit their jobs.

### Users

Users will currently come to Jobanalyzer via its command line interface (there is room for a web
interface or other GUI in the future).  The user will query the system for information about her own
running or completed jobs to examine them for system utilization and scalability.

There may be an interface that allows a job to be compared post-mortem with a database of benchmark
results for the purpose of finding a machine that is a good fit for the next run of the job.

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

THIS IS EVOLVING.

At least these:

- web-based load dashboard that shows current or near-current and historical loads of all known
  systems, specific UI TBD (see Use cases, below)

- web-based alert dashboard that shows alerts for various conditions that are of interest, such
  as inappropriate system use, dead processes, and downtime, specific triggers TBD (see Use cases,
  below), with functionality for extracting reports of users or jobs that regularly cause alerts

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

## User stories that will be supported

User stories are given identifier-like names so that they can more easily be referenced elsewhere.
*A* is a systems administrator.  *U* is a user.  *P* is a member of the public.  *D* is a decision
maker.  NAIC management is lumped in with the administrators, mostly.

Unless otherwise stated, we are assuming that there is continual data collection (sampling) on all
systems.

### "Admin" user stories

#### `adm_systems_at_a_glance`

*Story 1:* *A* wants a quick overview of the state of the various systems: whether they are up, and
how busy they are.

To retrieve these data, *A* goes to the web-based overview dashboard and selects the system(s) and
time window, and a number of plots will be shown, one per system, all on one page.  For multi-node
systems, all nodes are selected and aggregate data are shown.  The quantities plotted are
system-relative cpu, main memory, gpu, and gpu memory usage.  (TBD: queue length) If a system is
down during a period this is marked clearly and distinguished from zero load.

*A* can select time periods from at least the set last day, last week, or last month.

*Story 2:* Deadweight processes. TODO

#### `adm_historical_system_load`

*Story 1:* *A* wants to know the historical load on a system (single node or all nodes) to see how
busy it has been.  *A* is interested in short-term historical utilization, on the order of hours to
weeks, in a time window ending at the present time.  *A* is interested in the load as a percentage of
the system resources: cpu, main memory, gpu, and gpu memory.  (TBD: queue length)

To retrieve the load, *A* goes to the web-based load dashboard and selects the appropriate system and
time period, and a plot is produced showing the interesting quantities.

The quantities need not be 100% up to date - data that are a couple of hours old are acceptable.

*A* can select time periods from at least the set last day, last week, or last month.

On multi-node systems, *A* can select at least between one specific node or all nodes.

On systems with many nodes, or when there are many systems to select between, *A* can type in the
system name or part of it to quickly select the system.

*Story 2:* *A* finds something interesting in the plot and wants to communicate this to a coworker or
archive the situation for later.  *A* needs an artifact that can be used later.

To get this artifact, *A* can (in the worst case) take a screenshot of the plot.  That requires the
screenshot to contain the information about the time range and system selected.

Alternatively, *A* can ask for an image representing the situation (a little cleaner but basically
just the same as the screenshot), or *A* can copy the URL of the plot, provided this URL contains all
the necessary information and is not restricted to *A*'s use.

#### `adm_historical_uptime`

*Story 1:* *A* wants to know the historical uptime of a system (single node or all nodes).  *A* is
interested in longer-term historical uptime, going back a few months or a year, up to the present
time.  *A* is interested in reporting the uptime to somebody.  Uptime is defined as a time during
which a user could have run a job.  A system can thus be running without being "up" per se.

To retrieve the uptime data, *A* goes to the web-based uptime dashboard and selects the appropriate
system and time period, and a report is produced in a text area showing the interesting quantities.
*A* can copy the text into email or into a report.

The quantities need not be 100% up to date - data that are up to a day old are acceptable.

*A* can select time periods from at least the set last week, last month, last quarter, and last year.

On multi-node systems, *A* can select at least between one specific node or all nodes.

On systems with many nodes, or when there are many systems to select between, *A* can type in the
system name or part of it to quickly select the system.

#### `adm_unused_capacity`

*Story 1:* A course is held during a time period, placing a great deal of strain on Fox (say).
Meanwhile, Saga (say) has a lot of spare capacity.  *A* is interested in learning about this imbalance
so that he can tell instructors to move work to Saga.

*A* will primarily see this situation on the web-based alert dashboard, where a report that runs
(say) once a day detects this situation and sets a flag.  Once detected, the flag persists until the
situation has not been detected for some days. *A* will have to verify that the situation is still
the case before alerting instructors.

*Story 2:* *A* does not want to have to poll the dashboard every morning, and wants to be able to
register to receive an alert by email when the situation arises.

*A* can state this preference in his profile on the console.

#### `adm_misfits`

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

*Story 1 (gpuhog):* *U* is occupying and using a lot of GPU resources for a long time (on a system
without a batch queue).  This creates a situation where the work of one user prevents many other
users from getting work done.  *A* wants to move *U* to a system that has more GPUs and a batch
system.

*Story 2 (whale):* This is a generalization of *Story 1* but applies to all systems: *U* is too big
for the system - uses too much cpu, gpu, or memory, or for too long a time - and in the way of other
users.  And it may not be a matter of a single job that *U* is running, but the aggregate of those
jobs.  *A* wants to know this so that *U* can be moved to a larger system.

*Story 3 (cpuhog):* *U* is occupying and using a lot of CPU resources for a long time but not using
any GPU, on a system reserved for GPU computation.  This creates a situation where the work of one
user prevents many other users from getting work done.  *A* wants to move *U* to a system without
GPUs.

*Story 4 (confused):* This is a generalization of *Story 3* but pertains more to systems with batch
queues: *U* has asked for resources but are not using them.  *A* wants to alert *U* to this fact.


#### `adm_software_use`

TODO This is the #20 "what software is been used by users" ticket, and also #32 a bit I guess.


### "User" user stories

#### `usr_resource_use`

TODO (See `verify_gpu_use` and `verify_resource_use` in README.md)

#### `usr_scalability`

TODO (See `verify_scalability` and `thin_pipe` in README.md, and there is the #18 "data access bottleneck" and #58 "present inter-node communications volume" use cases)

### "Public and decision-maker" user stories

TBD
