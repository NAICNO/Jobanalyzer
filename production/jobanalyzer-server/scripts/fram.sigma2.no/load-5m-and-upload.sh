#!/usr/bin/bash

# Analysis job to run on one node every 5m.  This job generates the `minutely` load reports for the
# nodes and uploads them to the server.
#
# It also generates a file of hostnames and uploads that.

# We need globbing, stay away from -f
set -eu -o pipefail

cluster=fram.sigma2.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

abbrev=fram
tag="Fram Nodes"

$naicreport_dir/naicreport load \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      -report-dir $report_dir \
		      -with-downtime 5 \
		      -tag minutely \
		      -none

$naicreport_dir/naicreport at-a-glance \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      -state-dir $state_dir \
		      -tag "$tag" \
		      > $report_dir/$abbrev-at-a-glance.json

$naicreport_dir/naicreport hostnames \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      > $report_dir/$abbrev-hostnames.json

upload_files="$report_dir/*-minutely.json $report_dir/$abbrev-hostnames.json $report_dir/$abbrev-at-a-glance.json"
source $script_dir/upload-subr.sh
