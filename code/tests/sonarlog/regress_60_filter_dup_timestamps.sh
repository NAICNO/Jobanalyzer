# The input file has three records in a single stream, the first two have the same timestamp but are
# distinguishable, the second of the two should be filtered out when we clean.

output=$($SONALYZE parse --clean --fmt=csv,separator,job,cputime_sec -- regress_60_filter_dup_timestamps.csv)
CHECK parse_duplicate_timestamps '*
1249151,314
1249151,316' "$output"
