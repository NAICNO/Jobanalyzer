#!/bin/bash

# This script is currently run by cron at boot time (see sonar-server.cron) to start Jobanalyzer's
# `infiltrate` server, which receives data from all logging nodes.

# Currently this is set up for HTTPS.  For HTTP, change $infiltrate_port here or in the config file,
# and remove the -server-key and -server-cert arguments below.

sonar_dir=${sonar_dir:-$HOME/sonar}
source $sonar_dir/server-config

data_dir=$sonar_dir/data
mkdir -p $data_dir

pidfile=$sonar_dir/infiltrate.pid
rm -f $pidfile
if [[ $infiltrate_https == 1 ]]; then
    $sonar_dir/infiltrate \
        -data-path $data_dir \
        -auth-file $infiltrate_auth_file \
        -port $infiltrate_https_port \
        -server-key $infiltrate_https_key \
        -server-cert $infiltrate_https_cert &
    infiltrate_pid=$!
else
    $sonar_dir/infiltrate \
        -data-path $data_dir \
        -auth-file $infiltrate_auth_file \
        -port $infiltrate_http_port &
    infiltrate_pid=$!
fi

echo $infiltrate_pid > $pidfile
echo "INFILTRATE RUNNING, PID=$infiltrate_pid"
