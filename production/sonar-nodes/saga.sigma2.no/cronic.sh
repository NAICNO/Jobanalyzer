#!/bin/bash
#
# This is a shell script that runs sonar once a minute.  It is used instead of cron on system where
# we're allowed to run shell scripts even when not logged in, but not allowed to run cron jobs.

# Measuring once a minute to get good resolution.
# Upload window must be shorter than sleep window, 20 seconds is a good margin.

sleep_window=60
export upload_window=40

while true; do
    ./sonar-batchless.sh &
    sleep $sleep_window
done
