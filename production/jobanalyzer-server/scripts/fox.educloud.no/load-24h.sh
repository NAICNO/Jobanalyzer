#!/usr/bin/bash

# Analysis job to run on one node every 24h.  This job generates the monthly and quarterly load
# reports for the nodes.

set -euf -o pipefail

cluster=fox.educloud.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

$naicreport_dir/naicreport load \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      -report-dir $report_dir \
		      -with-downtime 5 \
		      -tag monthly \
		      -daily \
		      -from 30d

$naicreport_dir/naicreport load \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      -report-dir $report_dir \
		      -with-downtime 5 \
		      -tag quarterly \
		      -daily \
		      -from 90d
