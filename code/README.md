# What's here?

### Files

* `build.sh` - build all programs in release mode
* `run_tests.sh` - build programs in various configurations and run test cases

### Subdirectories

* `dashboard/` - HTML+CSS+JS source code for the web front-end
* `exfiltrate/` - Go source code for a program that ships Sonar data to a remote host, also see `infiltrate/`
* `go-utils/` - Go source code for utility functions used by all the Go programs in this repo
* `infiltrate/` - Go source code for a program that receives Sonar data on a host, also see `exfiltrate/`
* `naicreport/` - Go source code for a program that queries the Sonar data and generates reports
* `sonalyze/` - Rust source code for a program that queries the Sonar data
* `sonalyzed/` - Go source code for a simple HTTP server that runs Sonalyze on behalf of a remote client
* `sonard/` - Go source code for a utility that runs Sonar in the background with custom settings
* `sonarlog/` - Rust source code for a library that reads and cleans up Sonar data, used by `sonalyze/`
* `sysinfo/` - Go source code for a utility that extracts the system configuration of the host
* `tests/` - Test cases for everything
