# Output shall be sorted by ascending host name and for each host, by ascending time.
#
# This is a really basic test that checks that records can be parsed and that the output looks sane
# (and does not change over time).  It does not yet check that the output values are actually
# correct (because I've not checked the input yet).
#
# simple-input.csv taken from Saga login node activity, user names redacted.

output=$($SONALYZE top -- simple-input.csv)

CHECK top_smoketest_output \
      "HOST: login-1
  2024-06-28 00:01 ................................................................
  2024-06-28 00:02 ................................................................
HOST: login-3
  2024-06-28 00:01 o.OO............OO.O.....o......OOOo............o.O......O...O..
  2024-06-28 00:02 oo.O............OO.O...o.o......OOOO.............OO......O...O..
HOST: login-4
  2024-06-28 00:01 .....o.OoO.o...........O...Oo.........oo.O.....O.O.O...o.....OO.
  2024-06-28 00:02 ......oo.....o.o.....o.....OO..........OoO..oo.o.O..O..O...oOOO." \
      "$output"

