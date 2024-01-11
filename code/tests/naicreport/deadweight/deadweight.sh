statefile=deadweight-temp-state.csv
rm -f $statefile

# Populate a state from a deadweight report.  There will be one section of output for
# each line in the input.

output=$($NAICREPORT deadweight --state-file $statefile --summary -- deadweight1.csv | sort)
CHECK deadweight_from_empty_state \
      'jonaslsa,2529933,ml8
jonaslsa,516303,ml8
limeng,2710480,ml2
limeng,3485500,ml3
limeng,3818491,ml6
tobiaslo,2385640,ml7' \
      "$output"

# Now update the state from a later report.  Most of the lines will not contribute but there
# are two new jobs.  The explicit --to is necessary to prevent old record from being expired.

output=$($NAICREPORT deadweight --to 2023-09-30 --state-file $statefile --summary -- deadweight2.csv | sort)
CHECK deadweight_from_populated_state \
      'joachipo,469167,ml7
poyenyt,3959759,ml6' "$output"

# The 6 reports from the first run should have been integrated into the state

output=$(wc -l $statefile | awk '{ print $1 }')
CHECK deadweight_final_state 8 "$output"

# TODO: Many things, including
#
# - expiry from the state
# - stuff to do with time windows

rm -f $statefile
