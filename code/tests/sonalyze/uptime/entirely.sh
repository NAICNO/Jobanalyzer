# A host that is missing from the log is down the entire time
# There are no records for ML9 in the input.

output=$($SONALYZE uptime --from 2023-10-09 --to 2023-10-10 --interval 5 --fmt=csv,all --host 'ml[8-9]' --config-file hosts.json -- smoketest.csv)
CHECK uptime_host_missing \
      "host,ml8.hpc.uio.no,down,2023-10-09 00:00,2023-10-09 22:00
host,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15
gpu,ml8.hpc.uio.no,up,2023-10-09 22:00,2023-10-09 22:15
host,ml8.hpc.uio.no,down,2023-10-09 22:15,2023-10-10 23:59
host,ml9.hpc.uio.no,down,2023-10-09 00:00,2023-10-10 23:59" \
      "$output"
