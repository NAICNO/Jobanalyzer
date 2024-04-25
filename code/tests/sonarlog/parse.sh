# Various plain logfile parsing tests.

# General parsing tests.  Records should be seen in order read.  GPU sets should be printed in a
# predictable order, low-to-high card index.
output=$($SONALYZE parse --fmt=csv,user,rolledup,pid,job,localtime,gpus,cputime_sec -- parse-data.csv)
CHECK parse_file \
      "zabbix,5,0,4093,2023-06-26 16:00,none,7
root,0,4079,4079,2023-06-26 16:05,none,0
larsbent,0,1090,1249151,2023-06-26 16:10,\"3,5,6\",9212
larsbent,0,1090,1249151,2023-06-26 16:05,\"4,5,6\",8912
larsbent,0,1089,1249151,2023-06-26 16:05,\"1,4,6\",314
larsbent,0,1090,1249151,2023-06-26 16:15,\"2,5,6\",9362
larsbent,0,1090,1249152,2023-06-26 16:16,\"4,5,7\",9362" \
"$output"

# Same input but spread across two files.  This tests that read_logfiles
# in logtree.rs does its job.  The Go version reads in parallel, so we must sort
output=$($SONALYZE parse --fmt=csv,user,rolledup,pid,job,localtime,gpus,cputime_sec -- parse-data1.csv parse-data2.csv | sort)
CHECK parse_file_multi \
      "larsbent,0,1089,1249151,2023-06-26 16:05,\"1,4,6\",314
larsbent,0,1090,1249151,2023-06-26 16:05,\"4,5,6\",8912
larsbent,0,1090,1249151,2023-06-26 16:10,\"3,5,6\",9212
larsbent,0,1090,1249151,2023-06-26 16:15,\"2,5,6\",9362
larsbent,0,1090,1249152,2023-06-26 16:16,\"4,5,7\",9362
root,0,4079,4079,2023-06-26 16:05,none,0
zabbix,5,0,4093,2023-06-26 16:00,none,7" \
"$output"

# Record with "infinity" value for cpu_pct, bug #139, should be dropped without affecting other records.
# (There are other choices, but this works OK.)
output=$($SONALYZE parse --fmt=job,cpu_pct -- parse_infty.csv)
CHECK parse_infty "1685424,2
1685426,1" "$output"

# Make sure we can parse the "gpufail" field (a recent addition).
output=$($SONALYZE parse --fmt=job,gpu_status -- parse_gpufail.csv | sort)
CHECK parse_gpufail "1269178,1
1269179,0" "$output"

# Input file does not exist.  The Go and Rust versions have different capitalization of the error message.
output=$($SONALYZE parse -- no-such-file.csv 2>&1)
exitcode=$?
CHECK_ERR parse_no_file $exitcode "$output" "[Nn]o such file or directory"

# This file has four records, the second has a timestamp that is out of range and the fourth has a
# timestamp that is malformed.  We should be left with two.
output=$($SONALYZE parse --fmt=csv,job,user -- bad-timestamp.csv)
CHECK parse_bad_timestamps "4079,root
2288850,riccarsi" "$output"

# Same test, but untagged.
if [[ $($SONALYZE version) =~ untagged_sonar_data ]]; then
    output=$($SONALYZE parse --fmt=csv,job,user -- bad-timestamp-untagged.csv)
    CHECK parse_bad_timestamps_untagged "4079,root
2288850,riccarsi" "$output"
fi

# This file has three records, the second has a GPU set that is malformed, so the record is dropped.
output=$($SONALYZE parse --fmt=csv,job,user -- bad-gpuset.csv)
CHECK parse_bad_gpuset "4079,root
2288850,riccarsi" "$output"

# Same test, but untagged (and hence it has a gpu mask).  Only the rust version can parse these data.
if [[ $($SONALYZE version) =~ untagged_sonar_data && ! ( $($SONALYZE version) =~ short_untagged_sonar_data ) ]]; then
    output=$($SONALYZE parse --fmt=csv,job,user -- bad-gpumask-untagged.csv)
    CHECK parse_bad_gpumask "4079,root
2288850,riccarsi" "$output"
fi
