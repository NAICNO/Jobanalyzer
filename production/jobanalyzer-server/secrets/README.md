# naic-monitor secrets

## sonalyzed password file

This has any number of lines on the format `username:password` and permits remote queries of the
data.

Currently all usernames have access to all data.  This will change.

## exfiltrate/infiltrate password file

On the nodes (for `exfiltrate`), it is a single line on `username:password` format.

On the server (for `infiltrate`) it has any number of lines on the format `username:password`.

Currently all users can upload data for any node.  This will change.

## HTTPS certificates

### Exfiltrate/infiltrate certificates

`exfiltrate` and `infiltrate` now communicate over standard HTTPS via the nginx proxy on
naic-monitor.uio.no, and `exfiltrate` needs access to the certificate chain for that host.  This is
also installed in the web server.  Basically, the `naic-monitor.uio.no_fullchain.crt` file is copied
to the node's `secrets/` directory and is referenced from `sonar.sh`.