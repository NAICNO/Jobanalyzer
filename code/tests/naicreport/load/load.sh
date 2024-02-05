# -*- electric-indent-local-mode: nil -*-

# Very basic tests for `naicreport load`

# TODO: can we do something with mktmp?
# TODO: can we dump output on stdout with "-output-path -" say, or a different option?
OUTPUT_DIR=load-output-dir

rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR
$NAICREPORT load \
            -sonalyze $SONALYZE \
            -config-file load-smoketest-cfg.json \
            -output-path $OUTPUT_DIR \
            -tag test \
            -hourly -from 2023-10-09 -to 2023-10-10 \
            -with-downtime 5 \
            -- load-smoketest.csv
exitcode=$?
if (( $exitcode != 0 )); then
    FAIL load_smoketest_setup "Exit code: $exitcode"
fi

output=$(cat $OUTPUT_DIR/ml7*json)

# There should be 48 elements in all arrays because they cover two 24-hour periods.
# Let's check three of them.

expected_count=48
if [[ $output =~ '"rcpu"':'['([^]]*)']' ]]; then
    rcpu_items=${BASH_REMATCH[1]}
    count=$(echo $rcpu_items | sed 's/,/\n/g' | wc -l)
    CHECK load_smoketest_rcpu_count $expected_count $count
else
    FAIL load_smoketest_rcpu_count "$output"
fi
if [[ $output =~ '"labels"':'['([^]]*)']' ]]; then
    labels_items=${BASH_REMATCH[1]}
    count=$(echo $labels_items | sed 's/,/\n/g' | wc -l)
    CHECK load_smoketest_labels_count $expected_count $count
else
    FAIL load_smoketest_labels_count "$output"
fi
if [[ $output =~ '"downhost"':'['([^]]*)']' ]]; then
    downhost_items=${BASH_REMATCH[1]}
    count=$(echo $downhost_items | sed 's/,/\n/g' | wc -l)
    CHECK load_smoketest_downhost_count $expected_count $count
else
    FAIL load_smoketest_downhost_count "$output"
fi

# Check some informational items

t=1
if [[ $output =~ '"bucketing":"hourly"' ]]; then
    t=0
fi
CHECK load_smoketest_bucketing 0 "$t"

# Now check the contents of some of the arrays: specific values, and their positions.  Since the
# data start at 22:00 UTC on the first day and there is bucketing by hour there should be 22 zero
# items preceding the first data item, and since there are only data within that hour, the remaining
# entries after should all be zero.  This is true for rcpu/rgpu/rmem/rgpumem alike.

# TODO: Check eg rcpu

# For downhost, we have no information about the host status before the first record yet we have to
# produce data for it.  In principle, since we don't know, the assumption could be that the host is
# up or down - sonalyze currently decides "down".
#
# For downgpu, the calculation is in principle the same, but there is a wrinkle: GPU failure data
# are only produced while the CPU is up.  So the bit vector looks different.  Here we've set the
# gpufail flag on a record while the CPU is up and that is enough to cause that bit to be set in the
# downgpu vector.
#
# This all feels pretty ad-hoc and is definitely subject to change.

CHECK load_smoketest_downhost '1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1' "$downhost_items"

if [[ $output =~ '"downgpu"':'['([^]]*)']' ]]; then
    downgpu_items=${BASH_REMATCH[1]}
    CHECK load_smoketest_downgpu '0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0' "$downgpu_items"
else
    FAIL load_smoketest_downgpu "$output"
fi

rm -rf $OUTPUT_DIR

