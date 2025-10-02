# Test data here should have at least two hosts and two jobs per host, with at least two records per
# job on each host.  In some cross-host jobs the max and min times should not come come from the
# same host.

# Merge jobs within hosts but not across hosts.  This leaves the metadata untouched because it is
# per-host, not per-stream or per-job.

output=$($SONALYZE parse --merge-by-host-and-job --fmt=separator,host,job,localtime -- merge.csv)
CHECK parse_merge_by_host_and_job \
      '*
ml1.hpc.uio.no,4052478,2023-09-13 22:00
ml1.hpc.uio.no,4052478,2023-09-13 23:30
*
ml1.hpc.uio.no,3784499,2023-09-13 22:00
ml1.hpc.uio.no,3784499,2023-09-13 23:10
*
ml1.hpc.uio.no,2851773,2023-09-13 22:05
ml1.hpc.uio.no,2851773,2023-09-14 02:25
*
ml1.hpc.uio.no,3784760,2023-09-13 23:45
ml1.hpc.uio.no,3784760,2023-09-14 01:10
*
ml8.hpc.uio.no,2851773,2023-09-13 20:00
ml8.hpc.uio.no,2851773,2023-09-13 23:25
*
ml8.hpc.uio.no,4052478,2023-09-13 22:00
ml8.hpc.uio.no,4052478,2023-09-13 23:30
*
ml8.hpc.uio.no,3744442,2023-09-13 22:00
ml8.hpc.uio.no,3744442,2023-09-13 23:45
*
ml8.hpc.uio.no,3784499,2023-09-13 22:00
ml8.hpc.uio.no,3784499,2023-09-13 23:10
*
ml8.hpc.uio.no,3784760,2023-09-13 23:50
ml8.hpc.uio.no,3784760,2023-09-14 01:15' \
      "$output"

output=$($SONALYZE metadata --merge-by-host-and-job --bounds -- merge.csv)
CHECK metadata_merge_by_host_and_job \
      "ml1.hpc.uio.no,2023-09-13 22:00,2023-09-14 02:25
ml8.hpc.uio.no,2023-09-13 20:00,2023-09-14 01:15" \
      "$output"

# Then cross-host

# Records are merged if there are records with the same timestamp on multiple hosts.  This happens
# for 4052478 and 3784499, because a record on ml1 have the same timestamp as on ml8, and there are
# therefore two output records.  We verify that data are merged properly by looking at the sums of
# cputime_sec values.  The fields for cputime_sec have been set up as powers of 2, identifying the
# records.
#
# There are two jobs for 3744442 on ml8 because they are too far apart to be merged (I think)...

# TODO: clearly the merged host name is far from ideal here, issue #150.

output=$($SONALYZE metadata --merge-by-job --bounds -- merge.csv)
CHECK metadata_merge_by_job \
      "ml8.hpc.uio.no,2023-09-13 20:00,2023-09-14 01:15
\"ml[1,8].hpc.uio.no\",2023-09-13 20:00,2023-09-14 02:25" \
      "$output"

output=$($SONALYZE parse --merge-by-job --fmt=separator,host,job,localtime,cputime_sec -- merge.csv)
CHECK parse_merge_by_job \
      "*
ml8.hpc.uio.no,3744442,2023-09-13 22:00,1024
ml8.hpc.uio.no,3744442,2023-09-13 23:45,2048
*
\"ml[1,8].hpc.uio.no\",2851773,2023-09-13 20:00,1
\"ml[1,8].hpc.uio.no\",2851773,2023-09-13 22:05,64
\"ml[1,8].hpc.uio.no\",2851773,2023-09-13 23:25,2
\"ml[1,8].hpc.uio.no\",2851773,2023-09-14 02:25,128
*
\"ml[1,8].hpc.uio.no\",4052478,2023-09-13 22:00,260
\"ml[1,8].hpc.uio.no\",4052478,2023-09-13 23:30,520
*
\"ml[1,8].hpc.uio.no\",3784499,2023-09-13 22:00,4112
\"ml[1,8].hpc.uio.no\",3784499,2023-09-13 23:10,49152
*
\"ml[1,8].hpc.uio.no\",3784760,2023-09-13 23:45,32
\"ml[1,8].hpc.uio.no\",3784760,2023-09-13 23:50,8192
\"ml[1,8].hpc.uio.no\",3784760,2023-09-14 01:10,65536
\"ml[1,8].hpc.uio.no\",3784760,2023-09-14 01:15,131072" \
      "$output"
