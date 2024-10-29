# Relative test cases

## Newer stuff

Before running these, make a symlink or file `test-settings` in the current directory that contains
settings for the run, see eg settings-naic-monitor.uio.no.  That file will be included by the test
cases.

* The `X-print.sh` scripts test that two versions of `sonalyze X` print identical data, for various
  formatting arguments.  They are mostly used to test the new reflective printer.

## Older stuff

The rest are test cases that tested the Go version relative to the Rust version (or more generally, with
a little work, one version relative to another, independently of language).

The tests are normally run manually with run_tests.sh, which will run a host and user specific
runner that has settings for input data that make sense on the host/user.

These tests are pretty stale.
