#!/bin/bash

# This script is currently run by cron at boot time (see jobanalyzer.cron) to start Jobanalyzer's
# `infiltrate` server, which receives data from all logging nodes.

# Currently this is set up for HTTP.  Run it behind an HTTPS proxy.

sonar_dir=${sonar_dir:-$HOME/sonar}
source $sonar_dir/server-config

data_dir=$sonar_dir/data
mkdir -p $data_dir

pidfile=$sonar_dir/infiltrate.pid
$sonar_dir/infiltrate \
    -data-path $data_dir \
    -auth-file $infiltrate_auth_file \
    -match-user-and-cluster \
    -port $infiltrate_http_port &
infiltrate_pid=$!

echo $infiltrate_pid > $pidfile
echo "INFILTRATE RUNNING, PID=$infiltrate_pid"
