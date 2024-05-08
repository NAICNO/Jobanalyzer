#!/bin/bash
#
# Ad-hoc test runner for lth's laptop.  To adapt this, change the settings at the top.

set -e
export DATA=~/m/data
export FROM=2023-09-01
export TO=2023-09-30
export JOB=326742
export CONFIG=~/p/Jobanalyzer/production/jobanalyzer-server/scripts/mlx.hpc.uio.no/mlx.hpc.uio.no-config.json

source test-generic.sh
