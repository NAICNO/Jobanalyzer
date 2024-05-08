# To be included from test-HOSTNAME.sh

echo "perf1"
./perf1.sh $DATA $FROM $TO $JOB

echo "help"
./help.sh

echo "parse1"
./parse1.sh $DATA $FROM $TO

echo "parse2"
./parse2.sh $DATA $FROM $TO

echo "parse3"
./parse3.sh $DATA $FROM $TO

echo "parse4"
./parse4.sh $DATA $FROM $TO

echo "parse5"
./parse5.sh $DATA $FROM $TO parse5-$(hostname).sh

echo "parse6"
./parse6.sh $DATA $FROM $TO

echo "parse7"
./parse7.sh $DATA $FROM $TO

echo "uptime1"
./uptime1.sh $DATA $FROM $TO

echo "uptime2"
./uptime2.sh $DATA $FROM $TO

echo "uptime3"
./uptime3.sh $DATA $FROM $TO

echo "uptime4"
./uptime4.sh $DATA $FROM $TO

echo "uptime5"
./uptime5.sh $DATA $FROM $TO

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
