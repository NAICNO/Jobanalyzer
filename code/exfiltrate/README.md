# Exfiltrate

`exfiltrate` - data forwarder for on-node agents.

## How to run

See comment block at the beginning of exfiltrate.go for all information.

## Future developments

None expected.  This is basically a tiny subset of `curl` plus a random-wait feature.  It may be
replaced by a shell script, or by an mqtt client if we switch from http to mqtt.

## Past sins

This used to be a lot more complicated, but the complications did not pay off.

