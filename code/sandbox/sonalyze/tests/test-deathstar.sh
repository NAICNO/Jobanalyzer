#!/bin/bash
#
# Ad-hoc test runner for lth's dev system.  To adapt this, change the settings at the top.

set -e
export DATA=~/sonar/data/mlx.hpc.uio.no
export FROM=2024-01-01
export TO=2024-01-31
export JOB=2620514
export CONFIG=~/p/Jobanalyzer/production/jobanalyzer-server/scripts/mlx.hpc.uio.no/mlx.hpc.uio.no-config.json

source test-generic.sh
