# The input file has three records in a single stream, the first has an earlier timestamp than
# the second but a higher cputime_sec mark.  When we clean, the second record should be filtered
# out because its CPU utilization field becomes negative.  The third remains, and should
# participate in the computation of the cpu_util_pct field.
#
# Computed utilization for the first one should be 100 b/c that's what the cpu% field carries; for
# the second one, it is (910-310)/(60*10)=600/600=1.
#
# However there's bug 166: the filtering happens too late, so the utilization computed for this
# record is (910-200)/(60*5)=2.367.

output=$($SONALYZE parse --clean --fmt=csv,job,cputime_sec,cpu_util_pct -- regress_63_filter_negative_utilization.csv)
CHECK parse_duplicate_timestamps '*
1249151,310,100
1249151,910,1' "$output" 166
