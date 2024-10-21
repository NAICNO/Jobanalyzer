#!/usr/bin/bash

# Analysis job to run on the analysis host every 1h.  This job generates the hourly and daily load
# reports for the nodes.

set -euf -o pipefail

cluster=betzy.sigma2.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

$naicreport_dir/naicreport load \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      -with-downtime 5 \
		      -tag daily \
		      -hourly \
		      -report-dir $report_dir

$naicreport_dir/naicreport load \
		      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
		      -with-downtime 5 \
		      -tag weekly \
		      -hourly \
		      -from 7d \
		      -report-dir $report_dir
