# Output shall be sorted by ascending host name and for each host, by ascending time For ML8, the
# first cpu% is taken from the first record and the second is computed from delta time and delta
# cputime_sec.
#
# The weird formatting is the result of some very old code interacting poorly with much newer code,
# see comments in cmd/load/print.go.
output=$($SONALYZE load --fmt=json,host,cpu --compact --none -- json_output.csv)
CHECK json_output \
      '[{"system":{"hostname":"ml4.hpc.uio.no","description":"Unknown","gpucards":"0"},"records":[{"host":"ml4.hpc.uio.no","cpu":"58"}]
},{"system":{"hostname":"ml8.hpc.uio.no","description":"Unknown","gpucards":"0"},"records":[{"host":"ml8.hpc.uio.no","cpu":"18"},{"host":"ml8.hpc.uio.no","cpu":"231"}]
}]' \
      "$output"

output=$($SONALYZE load --fmt=json,host,cpu --compact --none -- empty_input.csv)
CHECK json_empty_output "[]" "$output"
