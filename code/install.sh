#!/bin/bash
#
# Script to build some things that need to be built and install them in ~/go/bin

set -o errexit

echo "======================================================================="
echo " GO SONALYZE RELEASE BUILD"
echo
( cd sonalyze ; go install )

echo "======================================================================="
echo " HEATMAP RELEASE BUILD"
echo
( cd heatmap ; go install )

echo "======================================================================="
echo "======================================================================="
echo "NORMAL COMPLETION"
