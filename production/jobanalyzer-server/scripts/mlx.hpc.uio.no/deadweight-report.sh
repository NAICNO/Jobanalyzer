#!/usr/bin/bash

# Meta-analysis job to run on one node every 12h.  This job prints a report on stdout, which will be
# emailed to the job owner by cron if nothing else is set up, and it generates a json file for
# upload.

set -euf -o pipefail

cluster=mlx.hpc.uio.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

# These update $state_dir/deadweight-state.csv; just nuke that file if you want to start the
# analysis from scratch.
#
# Typical running time per invocation on ML nodes: 10-20ms

$naicreport_dir/naicreport deadweight -state-dir $state_dir -from 4w > $report_dir/ml-deadweight-report.txt
if [[ -s $report_dir/ml-deadweight-report.txt ]]; then
    $naicreport_mail -s "ML deadweight report" "$ml_deadweight_recipient" < $report_dir/ml-deadweight-report.txt
fi
#$naicreport_dir/naicreport deadweight -state-dir $state_dir -from 4w -json > $report_dir/ml-deadweight-report.json
