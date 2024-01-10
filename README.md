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
implementation.  See [REQUIREMENTS.md](REQUIREMENTS.md) for requirements and a list of specific use
cases.


### Admins

Admins will come to Jobanalyzer via [its web interface](http://http://158.39.48.160/).  The current
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


## What are in the different subdirectories and files?

### Files

* `build.sh` - build all programs in release mode
* `DESIGN.md` - architecture and implementation overview
* `OLDER-USE-CASES.md` - original use cases
* `REQUIREMENTS.md` - cleaned-up requirements and use cases
* `run_tests.sh` - build programs in various configurations and run test cases

### Subdirectories

* `dashboard/` - HTML+CSS+JS source code for the web front-end
* `exfiltrate/` - Go source code for a program that ships Sonar data to a remote host, also see `infiltrate/`
* `go-utils/` - Go source code for utility functions used by all the Go programs in this repo
* `infiltrate/` - Go source code for a program that receives Sonar data on a host, also see `exfiltrate/`
* `naicreport/` - Go source code for a program that queries the Sonar data and generates reports
* `presentations/` - Slides for various presentations given
* `production/` - All sorts of files and scripts for running Sonar and Jobanalyzer in production
* `scripts/` - Ad-hoc reports implemented as shell scripts
* `sonalyze/` - Rust source code for a program that queries the Sonar data
* `sonalyzed/` - Go source code for a simple HTTP server that runs Sonalyze on behalf of a remote client
* `sonard/` - Go source code for a utility that runs Sonar in the background with custom settings
* `sonarlog/` - Rust source code for a library that reads and cleans up Sonar data, used by `sonalyze/`
* `sysinfo/` - Go source code for a utility that extracts the system configuration of the host
* `tests/` - Test cases for everything
