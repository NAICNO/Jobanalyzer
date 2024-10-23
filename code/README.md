# What's here?

### Make

Run `make build` (here or above) to make everything, `make clean` to clean it, `make test` to run
unit tests and linting.

Run `make regress` to run various more complicated tests and regression tests.

### Subdirectories

#### Core programs and libraries

* `dashboard/` - HTML+CSS+JS source code for the old web front-end
* `dashboard-2/` - HTML+CSS+React+TypeScript source code for the new web front-end
* `go-utils/` - Go source code for utility functions used by all the Go programs in this repo
* `naicreport/` - Go source code for a program that runs sonalyze and generates reports
* `sonalyze/` - Go source code for a program that ingests and queries the Sonar data

#### Utility programs, tests, and other

* `dashtest/` - Test code for the dashboard
* `heatmap/` - Go source code for a utility that parses text data and creates a simple 2D heat map of it
* `jsoncheck/` - Go source code for a simple utility that syntax checks JSON data
* `make-cluster-config/` - Go source code for a utility that collects sysinfo data into cluster config files
* `netsink/` - Go source code for a program that receives random input from a port and logs it
* `numdiff/` - Go source code for a program that does approximate number-aware file comparison
* `slurminfo/` - Go source code for a utility that runs `sinfo` and constructs a cluster config file
* `sonard/` - Go source code for a utility that runs Sonar in the background with custom settings
* `tests/` - Test cases for everything

#### Work in progress

* `sacctd/` - Go source code for a core program that runs Slurm's `sacct` on a cluster and massages the output for exfiltration

#### Attic - obsolete but not deleted

* `attic/exfiltrate/` - Go source code for a program that ships Sonar data to a remote host
* `attic/infiltrate/` - Go source code for a program that receives Sonar data on a server
* `attic/rustutils/` - Rust source code for utility functions used by all the Rust programs in this repo
* `attic/sonalyze/` - Rust source code for a program that queries the Sonar data
* `attic/sonalyzed/` - Go source code for an HTTP server that runs Sonalyze on behalf of a remote client
* `attic/sonarlog/` - Rust source code for a library that reads and cleans up Sonar data, used by `sonalyze/`
* `attic/sysinfo/` - Go source code for a utility that extracts the system configuration of the host
