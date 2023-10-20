#!/bin/bash
#
# Script to set up jobanalyzer on web server (tentative)
#
# Run this from the Jobanalyzer top directory.

echo "If you're sure you want to run this, remove the following line"
exit 1

WWWDIR=/var/www/html

mkdir -p $WWWDIR
mkdir -p $WWWDIR/output

cp dashboard/*.html dashboard/*.js dashboard/*.css $WWWDIR
