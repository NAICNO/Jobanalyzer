#!/usr/bin/env bash

# Meta-analysis job to run on one node every 12h.  This job prints a
# report on stdout, which will be emailed to the job owner by cron if
# nothing else is set up.

set -euf -o pipefail

sonar_dir=$HOME/sonar
sonar_data_dir=$sonar_dir/data
load_report_path=$sonar_data_dir/load-reports

mkdir -p ${load_report_path}

# These update $sonar_data_dir/cpuhog-state.csv; just nuke that file
# if you want to start the analysis from scratch.
#
# Typical running time per invocation on ML nodes: 10-20ms

$sonar_dir/naicreport ml-cpuhog -data-path $sonar_data_dir -from 4w
$sonar_dir/naicreport ml-cpuhog -data-path $sonar_data_dir -from 4w -json > $load_report_path/cpuhog-report.json
