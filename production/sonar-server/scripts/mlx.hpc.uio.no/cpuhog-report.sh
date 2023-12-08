#!/usr/bin/bash

# Meta-analysis job to run on one node every 12h.  This job prints a report on stdout, which will be
# emailed to the job owner by cron if nothing else is set up, and it generates a json file for
# upload.

set -euf -o pipefail

cluster=mlx.hpc.uio.no
sonar_dir=$HOME/sonar
report_dir=$sonar_dir/reports/$cluster
state_dir=$sonar_dir/state/$cluster

mkdir -p ${report_dir}

# These update $state_dir/cpuhog-state.csv; just nuke that file
# if you want to start the analysis from scratch.
#
# Typical running time per invocation on ML nodes: 10-20ms

$sonar_dir/naicreport ml-cpuhog -data-path $state_dir -from 4w
$sonar_dir/naicreport ml-cpuhog -data-path $state_dir -from 4w -json > $report_dir/ml-violator-report.json
