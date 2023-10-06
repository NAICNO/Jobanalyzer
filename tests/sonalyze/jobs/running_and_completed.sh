# This is still running because the record on ml4 is the last record observed on that host
output=$($SONALYZE jobs -u- --host ml4 --min-samples=1 -f 2023-10-04 --fmt=csv,host,job --running -- running_and_completed.csv)
CHECK running_ml4 "ml4,1269178" "$output"

# Regression test for Jobanalyzer#118.  The problem is that the metadata for the job log - earliest
# and latest records seen - are computed across hosts.  Based on `latest` in particular we compute
# the values used to filter by --running and --completed.  But a job J1 on host H1 is flagged as no
# longer running if there was a record R2 on host H2 that had a later timestamp than any record in
# J1.  This is wrong if hosts are independent (as they are on the ML nodes).
#
# The data for this test case therefore include records for two different hosts to create the above
# situation.  The job on ml8 ends before the job on ml4, but the former should still be marked as
# completed.

# This ends before the last record on that host and should be completed
output=$($SONALYZE jobs -u- --host ml8 --min-samples=1 -f 2023-10-04 --fmt=csv,host,job --completed -- running_and_completed.csv)
CHECK completed_ml8 "ml8,90548" "$output" "jobanalyzer#118"

# This ends at the same time as the last record on that host and should be running
output=$($SONALYZE jobs -u- --host ml8 --min-samples=1 -f 2023-10-04 --fmt=csv,host,job --running -- running_and_completed.csv)
CHECK running_ml8 "ml8,2092901" "$output" "jobanalyzer#118"


