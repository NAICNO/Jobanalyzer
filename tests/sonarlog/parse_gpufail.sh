#!/bin/bash

# Using `sonalyze jobs` isn't the most obvious thing but it works OK for now.  Once we have
# `sonalyze uptime` we can use that instead, maybe.  But the best way to do this would be some
# direct way of using sonalyze as a parser that just spits out its understanding of what it has
# read: effectively, a shell for running sonarlog.

export t_name=jobs/parse_gpufail_1
export t_expected=1269178,1
export t_output=$($SONALYZE jobs -u einarvid --min-samples=1 -f 2023-10-04 -t 2023-10-04 --fmt=csv,job,gpufail -- parse_gpufail.csv)
source ../harness.sh

export t_name=jobs/parse_gpufail_2
export t_expected=1269179,0
export t_output=$($SONALYZE jobs -u sinbad --min-samples=1 -f 2023-10-04 -t 2023-10-04 --fmt=csv,job,gpufail -- parse_gpufail.csv)
source ../harness.sh
