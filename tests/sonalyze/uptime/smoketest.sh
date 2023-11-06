# Basic tests for the uptime verb.

# This looks the way it does because, even though there is a gap in the timeline, the gap is
# slightly shorter than 10 minutes due to when the analysis ran.  See #136.
output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-10 --interval 5 --fmt=csv,all --host ml8 -- smoketest.csv)
CHECK uptime_smoketest8_all \
      "host,ml8.hpc.uio.no,down,2023-10-09 00:00,2023-10-09 22:00
host,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15
gpu,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15
host,ml8.hpc.uio.no,down,2023-10-09 22:15,2023-10-10 23:59" \
    "$output"

# This is like uptime_smoketest8_all but shortening the interval to 4 minutes (ie, 8 minutes in the
# program) reveals that the system was down for a short time.  See #136.
output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-10 --interval 4 --fmt=csv,all --host ml8 -- smoketest.csv)
CHECK uptime_smoketest8_4min \
      "host,ml8.hpc.uio.no,down,2023-10-09 00:00,2023-10-09 22:00
host,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:00
host,ml8.hpc.uio.no,down,2023-10-09 22:00,2023-10-09 22:10
gpu,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:00
host,ml8.hpc.uio.no,up,2023-10-09 22:10,2023-10-09 22:15
gpu,ml8.hpc.uio.no,up,2023-10-09 22:10,2023-10-09 22:15
host,ml8.hpc.uio.no,down,2023-10-09 22:15,2023-10-10 23:59" \
    "$output"

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-10 --interval 5 --fmt=csv,all --host ml8 --only-up -- smoketest.csv)
CHECK uptime_smoketest8_only_up \
      "host,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15
gpu,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15" \
    "$output"

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-10 --interval 5 --fmt=csv,all --host ml8 --only-down -- smoketest.csv)
CHECK uptime_smoketest8_only_down \
      "host,ml8.hpc.uio.no,down,2023-10-09 00:00,2023-10-09 22:00
host,ml8.hpc.uio.no,down,2023-10-09 22:15,2023-10-10 23:59" \
    "$output"

# Same as above
output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-10 --interval 5 --fmt=json,all --host ml8 --only-down -- smoketest.csv)
CHECK uptime_smoketest8_only_down_json \
      '[{"device":"host","host":"ml8.hpc.uio.no","state":"down","start":"2023-10-09 00:00","end":"2023-10-09 22:00"},{"device":"host","host":"ml8.hpc.uio.no","state":"down","start":"2023-10-09 22:15","end":"2023-10-10 23:59"}]' \
    "$output"

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-11 --interval 5 --fmt=csv,all --host ml7 -- smoketest.csv)
CHECK uptime_smoketest7_all \
      "host,ml7.hpc.uio.no,down,2023-10-09 00:00,2023-10-09 22:05
host,ml7.hpc.uio.no,up,2023-10-09 22:05,2023-10-09 22:25
gpu,ml7.hpc.uio.no,up,2023-10-09 22:05,2023-10-09 22:10
gpu,ml7.hpc.uio.no,down,2023-10-09 22:10,2023-10-09 22:15
gpu,ml7.hpc.uio.no,up,2023-10-09 22:15,2023-10-09 22:25
host,ml7.hpc.uio.no,down,2023-10-09 22:25,2023-10-11 23:59" \
      "$output"

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-11 --interval 5 --fmt=csv,all --host ml7 --only-down -- smoketest.csv)
CHECK uptime_smoketest7_only_down \
      "host,ml7.hpc.uio.no,down,2023-10-09 00:00,2023-10-09 22:05
gpu,ml7.hpc.uio.no,down,2023-10-09 22:10,2023-10-09 22:15
host,ml7.hpc.uio.no,down,2023-10-09 22:25,2023-10-11 23:59" \
      "$output"

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-11 --interval 5 --fmt=csv,all --host ml7 --only-up -- smoketest.csv)
CHECK uptime_smoketest7_only_up \
      "host,ml7.hpc.uio.no,up,2023-10-09 22:05,2023-10-09 22:25
gpu,ml7.hpc.uio.no,up,2023-10-09 22:05,2023-10-09 22:10
gpu,ml7.hpc.uio.no,up,2023-10-09 22:15,2023-10-09 22:25" \
      "$output"

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-11 --interval 5 --fmt=csv,all --host 'ml[7-9]' --only-up -- smoketest.csv)
CHECK uptime_smoketest_multi_only_up \
      "host,ml7.hpc.uio.no,up,2023-10-09 22:05,2023-10-09 22:25
gpu,ml7.hpc.uio.no,up,2023-10-09 22:05,2023-10-09 22:10
gpu,ml7.hpc.uio.no,up,2023-10-09 22:15,2023-10-09 22:25
host,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15
gpu,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15" \
      "$output"

output=$($SONALYZE uptime --interval 5 --fmt=json,all -- empty_input.csv)
CHECK uptime_json_empty "[]" "$output"
