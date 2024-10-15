#!/usr/bin/bash

# Analysis job to run on the analysis host every 1h.  This job generates the hourly and daily load
# reports for the nodes.

set -euf -o pipefail

cluster=fox.educloud.no
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

$naicreport_dir/naicreport load \
                      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
                      -tag fox-cpu-weekly \
                      -hourly \
                      -from 7d \
                      -group 'c1-[5-29]' \
                      -report-dir $report_dir

$naicreport_dir/naicreport load \
                      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
                      -tag fox-gpu-weekly \
                      -hourly \
                      -from 7d \
                      -group 'gpu-[1-10]' \
                      -report-dir $report_dir

$naicreport_dir/naicreport load \
                      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
                      -tag fox-int-weekly \
                      -hourly \
                      -from 7d \
                      -group 'int-[1-4]' \
                      -report-dir $report_dir

$naicreport_dir/naicreport load \
                      -sonalyze $naicreport_dir/sonalyze \
                      $data_source \
                      -tag fox-login-weekly \
                      -hourly \
                      -from 7d \
                      -group 'login-[1-4]' \
                      -report-dir $report_dir
