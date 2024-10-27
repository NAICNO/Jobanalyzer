# This file to be included from test-HOSTNAME.sh or other test runner.
#
# We need the following to be defined and exported by the script running this one:
#   OLD_SONALYZE is one binary, intended to be the original
#   NEW_SONALYZE is the other binary, intended to be the update
#   NUMDIFF is a program that can do field-wise numeric approximately-equal comparisons
#   DATA_PATH is the data dir, for --data-path
#   FROM is the starting date, for --from
#   TO is the ending date, for --to
#   JOB is a job number in the data, for --job
#
# There are also some per-test variables, see test-deathstar.sh for a canonical-ish definition.

echo "perf1"
./perf1.sh $DATA $FROM $TO $JOB

echo "help"
./help.sh

echo "parse1"
./parse1.sh

echo "parse2"
./parse2.sh

echo "parse5"
./parse5.sh

echo "parse6"
./parse6.sh

echo "parse7"
./parse7.sh

echo "uptime1"
./uptime1.sh

echo "uptime4"
./uptime4.sh

echo "uptime5"
./uptime5.sh

echo "load1"
./load1.sh $DATA $FROM $TO

echo "load2"
./load2.sh $DATA $FROM $TO

echo "jobs1"
./jobs1.sh $DATA $FROM $TO

echo "jobs2"
./jobs2.sh $DATA $FROM $TO

echo "profile1"
./profile1.sh $DATA $FROM $TO $JOB

echo "profile2"
./profile2.sh $DATA $FROM $TO $JOB

echo "profile3"
./profile3.sh $DATA $FROM $TO $JOB

echo "profile4"
./profile4.sh $DATA $FROM $TO $JOB

echo "done"
