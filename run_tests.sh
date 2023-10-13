#!/bin/bash
#
# This compiles all programs, runs all module tests, then runs all regression and blackbox tests.
#
# To just run the regression and blackbox tests: cd tests ; ./run_tests.sh

set -o errexit

# Run sonarlog/sonalyze whitebox tests.  I can't get `cargo test --all-features` to work, so test
# separate feature flags separately.
( cd sonarlog ; cargo test ; cargo test -F untagged_sonar_data )
( cd sonalyze ; cargo test ; cargo test -F untagged_sonar_data )

# Build sonalyze and test online help
( cd sonalyze ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd sonalyze ; target/debug/sonalyze jobs --fmt=help > /dev/null )

# Run naicreport whitebox tests
( cd naicreport/util ; go test )
( cd naicreport/joblog ; go test )
( cd naicreport/jobstate ; go test )
( cd naicreport/storage ; go test )

# Build naicreport and test online help
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )

# Run regression and blackbox tests in default configurations (built above)
( cd tests ; ./run_tests.sh )

# Run regression and blackbox tests in other configurations
( cd sonalyze ; cargo build -F untagged_sonar_data )
( cd tests ; ./run_tests.sh )

