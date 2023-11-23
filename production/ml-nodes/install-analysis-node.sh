#!/bin/bash
#
# Script to set up jobanalyzer on compute node (tentative)
#
# Run this from the Jobanalyzer top directory.

echo 'If you change $SONARDIR you will want to edit jobanalyzer-analysis.cron and *all*'
echo 'the shell scripts after they have been copied and before you run crontab.'
echo
echo
echo "If you're sure you want to run this, remove the following line"
exit 1

./build.sh

SONARDIR=~/sonar
mkdir -p $SONARDIR

cp sonalyze/release/sonalyze $SONARDIR
cp naicreport/naicreport $SONARDIR
cp production/ml-nodes/ml-nodes.json $SONARDIR
cp production/ml-nodes/jobanalyzer-analysis.cron $SONARDIR
cp production/ml-nodes/cpuhog*.sh \
   production/ml-nodes/deadweight*.sh \
   production/ml-nodes/upload*.sh \
   production/ml-nodes/webload*.sh \
   $SONARDIR

echo 'You need to update $SONARDIR/upload-config.sh with proper credentials and addresses'
exit 1

echo 'Even if you did not change $SONARDIR you must edit jobanalyzer-analysis.cron'
echo 'to set up the proper path (a regrettable wrinkle)'
exit 1

crontab -r
crontab $SONARDIR/jobanalyzer-analysis.cron
