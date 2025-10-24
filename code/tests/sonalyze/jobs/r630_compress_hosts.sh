# Regression test for issue #630 - jobs printer would wrongly recompress compressed hosts.
#
# Test data were obtained from production data by:
#
#   sonalyze parse -job 937627 -f 2d -fmt csvnamed,roundtrip,nodefaults
#     -data-dir ~/sonar/data/fox.educloud.no
#     -config-file ~/sonar/cluster-config/fox.educloud.no-config.json
#
# and the user name subsequently anonymised.

# This used to print "c1-[101918-,].fox" for the host name
output=$($SONALYZE jobs --user - -f 2024-10-21 --fmt=csv,host,user,job,cmd -config-file r630_conf.json -- r630_compress_hosts.csv)
CHECK r630_conf_merge '"c1-[10,18-19].fox",ec-testuser,937627,rosetta_scripts' "$output"

# Fixed output keeps only the node name not domain
output=$($SONALYZE jobs --user - -f 2024-10-21 --fmt=noheader,host,user,job,cmd -config-file r630_conf.json -- r630_compress_hosts.csv)
CHECK r630_conf_merge_fixed 'c1-[10,18-19]  ec-testuser  937627  rosetta_scripts' "$output"

# This used to print "c1-[101918-,].fox" for the host name
output=$($SONALYZE jobs --batch --user - -f 2024-10-21 --fmt=csv,host,user,job,cmd -- r630_compress_hosts.csv)
CHECK r630_batch_merge '"c1-[10,18-19].fox",ec-testuser,937627,rosetta_scripts' "$output"

# This was never broken but it's the third possible case.
#
# The extra sort is because there's no sorting by host name in the output from sonalyze at present.
# A sensible strategy would be to sort the records in field order, adapting to whatever spec is
# requested, or by specifying sort fields as an extra parameter.
output=$($SONALYZE jobs --user - -f 2024-10-21 --fmt=csv,host,user,job,cmd -- r630_compress_hosts_noepoch.csv | sort)
CHECK r630_no_merge 'c1-10.fox,ec-testuser,937627,rosetta_scripts
c1-18.fox,ec-testuser,937627,rosetta_scripts
c1-19.fox,ec-testuser,937627,rosetta_scripts' "$output"

output=$($SONALYZE jobs --user - -f 2024-10-21 --fmt=noheader,host,user,job,cmd -- r630_compress_hosts_noepoch.csv | sort)
CHECK r630_no_merge_fixed 'c1-10  ec-testuser  937627  rosetta_scripts
c1-18  ec-testuser  937627  rosetta_scripts
c1-19  ec-testuser  937627  rosetta_scripts' "$output"
