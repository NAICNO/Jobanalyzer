#!/bin/bash
#
# This is meant to run on naic-monitor, not remotely, but is easily adapted.

SONALYZE=${SONALYZE:-../../code/sonalyze/sonalyze}

$SONALYZE sacct -data-dir ~/sonalyze-test/data/fox.educloud.no -min-elapsed 7200 -min-reserved-mem 20 -min-reserved-cores 32 -f 3d -fmt awk,default \
    | ./heatmap -ppm \
    | pnmtopng > fox-2h-20G-32cpu-3d.png

$SONALYZE sacct -data-dir ~/sonalyze-test/data/fox.educloud.no -min-elapsed 3600 -min-reserved-mem 10 -min-reserved-cores 16 -f 3d -fmt awk,default \
    | ./heatmap -ppm \
    | pnmtopng > fox-1h-10G-16cpu-3d.png

