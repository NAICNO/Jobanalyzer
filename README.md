# Cross system Jobanalyzer

Jobanalyzer: Easy-to-use resource usage reporting and analyses.

## Overview

Jobanalyzer is a set of tools providing the following types of services:

- for admins: monitoring of current and historical utilization, as well as usage patterns
- for users: first-level analyses of computation patterns, with a view to appropriate system
  use and scalability - cpu use, gpu use, communication use

The tool set is expected to grow over time.

Current tools are based on a system sampler, and provide information based on collected samples.
See [doc/DESIGN.md](doc/DESIGN.md) for more information on the technical architecture and its
implementation.  See [doc/REQUIREMENTS.md](doc.REQUIREMENTS.md) for requirements and a list of specific use
cases.


### Admins

Admins will come to Jobanalyzer via [its web interface](http://naic-report.uio.no).  The current
interface is bare-bones and consists of a cluster- and node-centric load dashboard, allowing the
current and historical load of each cluster and node to be examined, along with some reports of
programs that mis-use the systems.  The UiO ML nodes and the UiO Fox supercomputer are currently
represented.

Data for the web interface are produced firstly by periodic analysis by the low-level `sonalyze`
tool and secondly by the higher-level `naicreport` tool, and the results of these analyses are
uploaded periodically to the web server.

The web interface will be extended with more functional dashboards, including alerts for actionable
items.  (Currently those alerts are emailed.)


### Users

Users will currently come to Jobanalyzer via its command line interface (there is room here for a
web interface or other GUI).  The primary interface is via the low-level `sonalyze` tool.  This tool
can be hard to use effectively, but does serve many use cases as described elsewhere in the
documentation.


## Setup

Jobanalyzer is a collective of programs running on three systems: compute nodes run `sonar` and
`sysinfo` to collect data; analysis nodes run `sonalyze`, `naicreport` and various other programs
and scripts to produce reports; and web nodes run a web server serving HTML, JS, and data.

See `production/README.md` for instructions about how to set everything up.


## What are in the different subdirectories?

* `adhoc-reports/` - Ad-hoc reports implemented as shell scripts or similar
* `code/` - Source code and tests for all programs
* `doc/` - design documents of various sorts, presentations, etc
* `production/` - All sorts of files and scripts for running Sonar and Jobanalyzer in production
