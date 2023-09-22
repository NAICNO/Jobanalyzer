#!/usr/bin/env bash

# Meta-analysis job to run on one node every 12h.  This job prints a
# report on stdout, which will be emailed to the job owner by cron if
# nothing else is set up.

set -euf -o pipefail

sonar_dir=$HOME/sonar
sonar_data_dir=$sonar_dir/data

# This updates $sonar_data_dir/deadweight-state.csv; just nuke that file
# if you want to start the analysis from scratch.
#
# Typical running time on ML nodes: 10-20ms

$sonar_dir/naicreport ml-deadweight -data-path $sonar_data_dir -from 2w
