#!/bin/bash
#
# Script to set up jobanalyzer (tentative)
#
# This isn't really meant to be run as-is.  It's a recipe for what to do.

echo "Not really meant to be run; instead edit it and run different pieces on different platforms"
exit 1

SONARDIR=~/sonar
WWWDIR=/var/www/html

# First build all executables in release mode
./build.sh
( cd ../sonar ; cargo build --release )

# Set up the back-end
mkdir -p $SONARDIR
mkdir -p $SONARDIR/data

cp ../sonar/target/release/sonar $SONARDIR
cp sonalyze/release/sonalyze $SONARDIR
cp naicreport/naicreport $SONARDIR
cp loginfo/loginfo $SONARDIR
cp production/ml-nodes/ml-nodes.json production/ml-nodes/*.sh production/ml-nodes/*.cron $SONARDIR
cp key-to-web-server.pem $SONARDIR

echo "You need to edit upload-data.sh and webload-5m-and-upload.sh to point to the key you just copied"

# Set up the front-end; this may be on a different host though
mkdir -p $WWWDIR
mkdir -p $WWWDIR/output

cp dashboard/*.html dashboard/*.js dashboard/*.css $WWWDIR

# As for the rest...

echo "On compute nodes you must setup cron or similar to run sonar:"
echo "  cd $SONARDIR ; crontab jobanalyzer.cron"
echo ""
echo "On the analysis node you must setup cron or similar to run the analysis:"
echo "  cd $SONARDIR"
echo "  vi jobanalyzer-moneypenny.cron"
echo ""
echo "Now change the definition of HOME and save the file.  Then:"
echo "  cd $SONARDIR ; crontab jobanalyzer-moneypenny.cron"
echo ""
echo "You also need to have a running webserver"
