#!/bin/bash
#
# Ad-hoc test runner for lth's dev system.

set -e

# Binaries
export OLD_SONALYZE=$(pwd)/../../attic/sonalyze/target/release/sonalyze
export OLD_NAME=rust
export NEW_SONALYZE=$(pwd)/../sonalyze
export NEW_NAME=go
export NUMDIFF=$(pwd)/../../numdiff/numdiff

# Program arguments
export DATA_PATH=~/sonar/data/mlx.hpc.uio.no
export FROM=2024-01-01
export TO=2024-01-31
export JOB=2620514
export CONFIG=$(pwd)/../../../production/jobanalyzer-server/scripts/mlx.hpc.uio.no/mlx.hpc.uio.no-config.json

# Settings for various tests.

export UPTIME5_CONFIG=$(pwd)/uptime5.cfg
export UPTIME5_HOST='ml5\.hpc\.uio\.no'

declare -A PARSE5_FILTER
export PARSE5_FILTER
PARSE5_FILTER["--host"]="ml3 ml7"
PARSE5_FILTER["--user"]="hermanno karths"
PARSE5_FILTER["--exclude-user"]="karths mateuwa"
PARSE5_FILTER["--command"]="python kited"
PARSE5_FILTER["--exclude-command"]="python kited"
PARSE5_FILTER["--job"]="2036281 2396512"
PARSE5_FILTER["--exclude-job"]="2036281 2396512"

source test-generic.sh
