#!/bin/bash
#
# Script to set up jobanalyzer on compute node (tentative)
#
# Run this from the Jobanalyzer top directory.

echo 'If you change $SONARDIR you will want to edit jobanalyzer.cron and sonar.sh'
echo 'after they have been copied and before you run crontab.'
echo
echo
echo "If you're sure you want to run this, remove the following line"
exit 1

# Sonar source code must be checked out in companion directory

( cd ../sonar ; cargo build --release ; cp target/release/sonar $SONARDIR )

SONARDIR=~/sonar
mkdir -p $SONARDIR

cp production/ml-nodes/sonar.sh $SONARDIR
cp production/ml-nodes/jobanalyzer.cron $SONARDIR

crontab -r
crontab $SONARDIR/jobanalyzer.cron

