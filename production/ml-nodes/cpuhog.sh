#!/usr/bin/env bash
#
# Run sonalyze for the `cpuhog` use case and capture its output in a file appropriate for the
# current time and system.

sonar_dir=$HOME/sonar
sonar_data_dir=$sonar_dir/data

year=$(date +'%Y')
month=$(date +'%m')
day=$(date +'%d')

output_directory=${sonar_data_dir}/${year}/${month}/${day}
mkdir -p ${output_directory}

# Report jobs that have used "a lot" of CPU and have run for at least 10 minutes but have not touched the
# GPU.  Reports go to stdout.
#
# What's "a lot" of CPU?  We define this for now as a peak of at least 10 cores.  This is imperfect
# but at least not completely wrong.
#
# The report is run on the data for the last 24h.  This should therefore be run about once every 12h
# at least, but ==> IMPORTANTLY, it MUST be run often enough that job IDs are not reused between
# consecutive runs.  It is not expensive, and can be run fairly often.

SONAR_ROOT=$sonar_data_dir $sonar_dir/sonalyze jobs --config-file=$sonar_dir/ml-nodes.json -u -  "$@" --no-gpu --min-rcpu-peak=10 --min-runtime=10m --fmt=csvnamed,tag:cpuhog,now,std,cpu-peak,gpu-peak,rcpu,rmem,start,end,cmd >> ${output_directory}/cpuhog.csv

