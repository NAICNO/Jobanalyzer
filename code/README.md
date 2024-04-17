# What's here?

### Files

* `build.sh` - build all programs in release mode
* `run_tests.sh` - build programs in various configurations and run test cases

### Subdirectories

#### Core programs and libraries

* `dashboard/` - HTML+CSS+JS source code for the web front-end
* `go-utils/` - Go source code for utility functions used by all the Go programs in this repo
* `infiltrate/` - Go source code for a program that receives Sonar data on a server
* `naicreport/` - Go source code for a program that queries the Sonar data and generates reports
* `rustutils/` - Rust source code for utility functions used by all the Rust programs in this repo
* `sonalyze/` - Rust source code for a program that queries the Sonar data
* `sonalyzed/` - Go source code for a simple HTTP server that runs Sonalyze on behalf of a remote client
* `sonarlog/` - Rust source code for a library that reads and cleans up Sonar data, used by `sonalyze/`

#### Utility programs, tests, and other

* `dashtest/` - Test code for the dashboard
* `exfiltrate/` - Go source code for an OBSOLETE program that ships Sonar data to a remote host
* `jsoncheck/` - Go source code for a simple utility that syntax checks JSON data
* `make-cluster-config/` - Go source code for a utility that collects sysinfo data into cluster config files
* `netsink/` - Go source code for a program that receives random input from a port and logs it
* `numdiff/` - Go source code for a program that does approximate number-aware file comparison
* `slurminfo/` - Go source code for a utility that runs `sinfo` and constructs a cluster config file
* `sonard/` - Go source code for a utility that runs Sonar in the background with custom settings
* `sysinfo/` - Go source code for an OBSOLETE utility that extracts the system configuration of the host
* `tests/` - Test cases for everything
* `sandbox/` - Playground for evolving code, experiments, etc
