#!/usr/bin/env bash
#
# Run sonalyze for the `deadweight` use case and capture its output in a file appropriate for the
# current time and system.

sonar_dir=$HOME/sonar
sonar_data_dir=$sonar_dir/data

year=$(date +'%Y')
month=$(date +'%m')
day=$(date +'%d')

output_directory=${sonar_data_dir}/${year}/${month}/${day}
mkdir -p ${output_directory}

# Report jobs that are not doing anything useful but are hanging onto system resources (zombies,
# defuncts, and maybe others).  This is defined entirely by the sonalyze `--zombie` option, for now.
#
# The report is run on the data for the last 24h.  This should therefore be run about once every 12h
# at least, but ==> IMPORTANTLY, it MUST be run often enough that job IDs are not reused between
# consecutive runs.  It is not expensive, and can be run fairly often.

SONAR_ROOT=$sonar_data_dir $sonar_dir/sonalyze jobs -u - "$@" --zombie --fmt=csvnamed,tag:deadweight,now,std,start,end,cmd >> ${output_directory}/deadweight.csv
