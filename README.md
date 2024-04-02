# Cross system Jobanalyzer

Jobanalyzer: Easy-to-use resource usage reporting and post-hoc job analyses.

## Overview

Jobanalyzer is a set of tools providing the following types of services:

- for admins: monitoring of current and historical utilization, as well as usage patterns
- for users: first-level analyses of computation patterns, with a view to appropriate system
  use and scalability - cpu use, gpu use, communication use

The tool set is expected to grow over time.

Current tools are based on a system sampler, and provide information based on collected samples.
See [doc/DESIGN.md](doc/DESIGN.md) for more information on the technical architecture and its
implementation.  See [doc/REQUIREMENTS.md](doc/REQUIREMENTS.md) for requirements and a list of
specific use cases.

The central feature of Jobanalyzer is that it aggregates and analyzes historical data, making it
possible to look into the past to examine both systems and individual jobs - more "what happened?"
than "what's going on?".  That said, up-to-date data are also available and current system and job
status are exposed in useful ways.


### Admins

Admins will mostly come to Jobanalyzer via [its web interface](http://naic-monitor.uio.no).  The
current interface is simple and consists of a cluster- and node-centric load dashboard (allowing the
current and historical load of each cluster and node to be examined, along with some example reports
of programs that mis-use the systems) and an interactive job querying and profiling facility.  The
UiO ML nodes, the UiO "Fox" supercomputer, and the Sigma2 "Saga" supercomputer are currently
represented.

Data for the web interface dashboard are produced firstly by periodic analyses by the low-level
`sonalyze` tool, secondly by the higher-level `naicreport` tool and some ad-hoc reports.  The
results of these analyses are uploaded periodically to the web server.

Data for the interactive queries are produced on demand.

The web interface will be extended with more functional dashboards, including alerts for actionable
items.  (Currently those alerts are emailed.)


### Users

Users will currently come to Jobanalyzer via its web interface for job query and profiling, or will
use it from the command line (the primary interface being the low-level `sonalyze` tool, which can
be run remotely).  These tools can be hard to use effectively, and need to be improved, but do serve
many use cases as described elsewhere in the documentation.


## Setup

Jobanalyzer is a collective of programs running on two groups of systems: compute nodes run `sonar`
collect data; while analysis nodes run `sonalyze`, `naicreport` and various other programs and
scripts to produce reports, as well as a web server serving HTML, JS, and data to remote clients.

Sample data are collected within a "cluster" and the data for a cluster has to be located on a
single analysis node, but beyond that there can be different analysis nodes for different clusters.

See `production/README.md` for instructions about how to set everything up.


## What are in the different subdirectories?

* `adhoc-reports/` - Ad-hoc reports implemented as shell scripts or similar
* `code/` - Source code and tests for all programs
* `doc/` - design documents of various sorts, presentations, etc
* `production/` - All sorts of files and scripts for running Sonar and Jobanalyzer in production
