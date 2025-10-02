# What's here?

### Make

Run `make build` (here or above) to make everything, `make clean` to clean it, `make test` to run
unit tests and linting, `make fmt` to format formattable things.

Run `make regress` to run various more complicated tests and regression tests.

### Subdirectories

#### Core programs and libraries

* `dashboard/` - HTML+CSS+JS source code for the old web front-end
* `dashboard-2/` - HTML+CSS+React+TypeScript source code for the new web front-end
* `go-utils/` - Go source code for utility functions used by all the Go programs in this repo
* `naicreport/` - Go source code for a program that runs sonalyze and generates reports
* `sonalyze/` - Go source code for a program that ingests and queries the Sonar data

#### Utility programs, tests, and other

* `heatmap/` - Go source code for a utility that parses text data and creates a simple 2D heat map of it
* `jsoncheck/` - Go source code for a simple utility that syntax checks JSON data
* `make-cluster-config/` - Go source code for a utility that collects sysinfo data into cluster config files
* `netsink/` - Go source code for a program that receives random input from a port and logs it
* `numdiff/` - Go source code for a program that does approximate number-aware file comparison
* `slurminfo/` - Go source code for a utility that runs `sinfo` and constructs a cluster config file
* `sonard/` - Go source code for a utility that runs Sonar in the background with custom settings
* `tests/` - Regression test cases for all the Go components
