LOGFILE=sonard-log.csv
VERBOSEFILE=verbose-log.txt
rm -f $LOGFILE $VERBOSEFILE

echo '`sonard` test takes 10s, please wait...'
( $SONARD -v -i 5 -s ./dummy-sonar.bash $LOGFILE 2> $VERBOSEFILE ) &
sleep 9
pkill sonard

output=$(grep "Running " $VERBOSEFILE | wc -l)
CHECK sonard_smoketest1 2 "$output"

output=$(grep "I am sonar" $LOGFILE | wc -l)
CHECK sonard_smoketest3 2 "$output"

rm -f $LOGFILE $VERBOSEFILE
