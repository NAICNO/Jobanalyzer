# naicreport

## Overview

`naicreport` is a shim around a number of commands that produce various kinds of reports and
maintain various kinds of state: `deadweight`, `glance`, `hostnames`, `load`, `mlcpuhog`.  The
functionality and usage of each command is documented in a block comment at the beginning of its
main source file, eg in `deadweight/deadweight.go`.

Command line parsing in the commands follows the Go standard, so:

- `-h` will get you brief help in most situations (consult the documentation for more information,
  as mentioned above)
- option names can use single dashes `-option` or double dashes `--option`
- option values can be stated with `-option value` or `-option=value`
- single-letter option names *can not* be merged with their values, as in `-f5d`

## Warning

Several of these commands have state, which is updated as necessary.  As a general rule, `naicreport`
does not have *thread-safe* storage, and the program should only be run on one system at a time.
