# Sonalyze

Sonalyze is an aggregation and query front-end for monitoring data (including
[Sonar](https://github.com/NordicHPC/Sonar) data, hence the name).  Its job is to make sense of the
data in various ways, perform selections on the data, and present results in useful form (both
human-readable and machine-readable).  By and large, other components in the ecosystem should go via
Sonalyze to access the monitoring data.

See [doc/MANUAL.md](doc/MANUAL.md) for most user information.

See [doc/README.md](doc/README.md) for an overview of all Sonalyze documentation.

## Subdirectories and files

* `application/` - main application logic, can be shared with other applications
* `cmd/` - all the application verbs except `daemon`
* `common/` - shared utility code
* `daemon/` - logic for the daemon
* `data/` - storage manager, top part: queries and cleans data coming from the database
* `db/` - storage manager, bottom part: storage interface
* `doc/` - all documentation
* `sonalyze.go` - command-line interface for sonalyze + daemon management
* `table/` - helper logic to define the "tables" produced by all the commands
* `util/` - misc utility scripts

In addition, Sonalyze uses a lot of code from the libraries in `../go-utils/`.
