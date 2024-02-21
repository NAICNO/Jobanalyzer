statefile=cpuhog-temp-state.csv
rm -f $statefile

# Populate a state from a cpuhog report.  There will be one section of output for
# each line in the input.

output=$($NAICREPORT ml-cpuhog --state-file $statefile --summary -- cpuhog1.csv | sort)
CHECK cpuhog_from_empty_state \
      'daniehh,2974731,ml6,5
einarvid,1017602,ml7,38
einarvid,1598971,ml1,34
einarvid,3971206,ml8,52
pubuduss,15306,ml9,31
tsauren,471142,ml8,167' \
      "$output"

# Now update the state from a later report.  Most of the lines will not contribute but there
# are two new jobs.  The explicit --to is necessary to prevent old record from being expired.

output=$($NAICREPORT ml-cpuhog --to 2023-10-25 --state-file $statefile --summary -- cpuhog2.csv | sort)
CHECK cpuhog_from_populated_state \
      'einarvid,2826710,ml2,32
tsauren,103737,ml8,160' "$output"

# The 6 reports from the first run should have been integrated into the state

output=$(wc -l $statefile | awk '{ print $1 }')
CHECK cpuhog_final_state 8 "$output"

# TODO: Many things, including
#
# - expiry from the state
# - stuff to do with time windows

rm -f $statefile

output=$($NAICREPORT ml-cpuhog --state-file $statefile --summary --json -- cpuhog1.csv | sort)
CHECK cpuhog_regression_220_part1 "1" "$(echo $output | grep pubuduss | wc -l | tr -d ' ')"

output=$($NAICREPORT ml-cpuhog --to 2023-10-25 --state-file $statefile --summary --json -- cpuhog2.csv | sort)
CHECK cpuhog_regression_220_part2 "1" "$(echo $output | grep tsauren | wc -l | tr -d ' ')"

rm -f $statefile
