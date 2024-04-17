# Important to remember that at this point, all times are UTC.  So a YYYY-MM-DD spec relates to
# midnight of that day UTC, not local time.  Note logfiles usually have local times including TZO.
#
# The time range of -f starts at 00:00Z on that date; the time range of -t ends just after 23:59:59Z
# on that date, but before 00:00Z on the next date.

# Should select records from this day only, UTC.  Then the endpoints are computed from those
# records, so the job should be perceived as running from midnight to midnight that day.  The file
# has data for the same job outside that date range.

output=$($SONALYZE jobs --user - --min-samples=1 -f 2023-10-03 -t 2023-10-03 --fmt=csv,job,start,end -- select_from_to.csv)
CHECK exact_range "3485500,2023-10-03 00:00,2023-10-03 23:55" "$output"
