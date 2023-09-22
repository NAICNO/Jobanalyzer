# naicreport

## Overview

`naicreport` is a component that ingests summaries produced by `sonalyze` and produces reports.  It
is separate from `sonalyze` because it has state: it knows what it's reported previously and will avoid
redundant reporting.

`naicreport` is really a superstructure for a number of more or less related and more or less ad-hoc
commands, expressed as verb arguments to `naicreport` itself, here are the ones considered so far:

- `naicreport ml-cpuhog <options>` will digest the `cpuhog.csv` logs produced by the
  `../production/sonalyze/ml-nodes/cpuhog.sh` script and will report new offending jobs to a Proper
  Authority.

- `naicreport ml-deadweight <options>` will digest the `deadweight.csv` logs produced by the
  `../production/sonalyze/ml-nodes/deadweight.sh` script and will report new offending processes to a
  Proper Authority.

- `naicreport ml-webload <options>` will (for now) invoke `sonalyze` on the `sonar` logs and will
  produce a system load report in a format digestable by the web dashboard.

Most of these commands have state, which is updated as necessary.  As a general rule, `naicreport`
does not have *thread-safe* storage, and the program should only be run on one system at a time.

Each command is implemented in a separate subdirectory, with shared code in `storage/` and `util/`.

## Design & implementation

`naicreport` is written in Go: the world wants static types, strong types, and garbage collection.
(Both `sonar` and `sonalyze` are written in Rust.  For `sonar` this was both fine and reasonable,
while for `sonalyze` it was not the most natural choice, but was instead driven by the prospect of
sharing the increasingly complicated log processing code with eg `jobgraph`.  `naicreport` does not
have that constraint.)

The data produced by `sonalyze` is always in "free csv form", ie using CSV syntax but with fields
tagged by field names and in arbitrary order, and different rows may not have the same number of
fields.  This allows its output to evolve, but it means `naicreport` must be a little flexible wrt
what it does when fields in its input data are missing.

