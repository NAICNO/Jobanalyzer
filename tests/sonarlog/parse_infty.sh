# Record with "infinity" value for cpu_pct, bug #139, should be dropped without affecting other records.
# (There are other choices, but this works OK.)
output=$($SONALYZE parse --fmt=job,cpu_pct -- parse_infty.csv)
CHECK parse_infty "1685424,2
1685426,1" "$output"
