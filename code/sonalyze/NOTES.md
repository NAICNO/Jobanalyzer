## TODO

Even though the various output formatters are stable there are some differences, defaults are
different (and may be buggy), and some verbs have custom formatting.  So it may be worth checking
all or several of the output formats with all the verbs.

* base library
  - [x] missing date truncate/add functionality, crucial
  - [x] test cases would be great
* sonarlog
  - [x] code complete
  - [x] unit testing (most of it)
* sonalyze & command infra
  - [x] local code complete modulo bugs
  - [x] bug: formatter help must sort data
  - [x] overall local test cases (help text mostly)
  - [x] local tests pass
  - [x] test more source/record filters (via parse)
* jobs
  - [x] sketch, does some things
  - [x] code complete
  - [x] tests cases for all options, bitwise equivalent to rust code
  - [x] test cases for formatting, ditto (all fields, some require config)
* load
  - [x] code complete
  - [x] depends on merging code in sonarlog
  - [x] test cases for all bucketing options
  - [x] test cases for all/last/compact
  - [x] test cases for formatting (all fields, some require config)
* metadata
  - [x] code complete (evolving as needs dictate)
  - [x] more tests
* parse
  - [x] code complete
  - [x] depends on merging code in sonarlog
  - [x] bug: output of plain command is sorted differently
  - [x] test cases for all arguments (will test merging)
  - [x] tests pass
  - [x] test formatting (most fields)
  - [x] test formatting of fields that are noisy now that we have numdiff
* profile
  - [x] code complete
  - [x] test cases
  - [x] tests pass
  - [x] test formatting (all fields)
* uptime
  - [x] code complete
  - [x] test cases exist
  - [x] test cases pass
  - [x] test with config file that has host that is not in data records
  - [x] test formatting (all fields, some require config)
  - [x] must sort by status too, b/c some records are otherwise indistinguishable
* testing (in addition to per-verb testing)
  - [x] regression test suite passes
  - [x] performance tests on a variety of data, for all commands
  - [ ] for both go and rust, self-test no-config vs config -- this is instructive...
  - [x] make sure scripts are sensibly configurable (esp with config file)
* remoting
  - [x] code complete
  - [x] test that local runs can be used to access remote functionality, and that all parameters are passed
  - [x] test that sonalyzed can interact with our version properly, in particular, that it does not use any short options
* transition
  - [x] run regression tests against new code again, then freeze
  - [x] move code/sonalyze/, code/sonarlog/, code/rustutils into code/attic
  - [x] update build.sh to build that code in code/attic (for now)
  - [x] update run_tests.sh to not test that code
  - [x] move code/sandbox/sonalyze into code
  - [x] include it in builds
  - [x] include it in regression tests
  - [x] update the sonalyze relative tests to work with the code in attic
  - [ ] remove the rest of code/sandbox (maybe branch, remove on current, later we rebase the branch to main?)
