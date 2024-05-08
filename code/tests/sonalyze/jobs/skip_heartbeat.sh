# If the heartbeat record is not properly skipped there will be multiple lines of output.

output=$($SONALYZE jobs --user - --min-samples=1 -f 2023-10-05 -t 2023-10-05 --fmt=csv,job -- skip_heartbeat.csv)
CHECK skip_heartbeat 1608145 "$output"
