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

The certificates for exfiltrate/infiltrate were generated like this:

```
openssl genrsa -out exfil-ca.key 2048
openssl req -new -x509 -days 3650 -key exfil-ca.key -subj "/C=IN/ST=KA/L=BL/O=Norwegian AI Cloud/CN=NAIC Root CA" -out exfil-ca.crt
openssl req -newkey rsa:2048 -nodes -keyout exfil-server.key -subj "/C=IN/ST=KA/L=BL/O=Norwegian AI Cloud/CN=naic-monitor.uio.no" -out exfil-server.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:naic-monitor.uio.no") -days 3650 -in exfil-server.csr -CA exfil-ca.crt -CAkey exfil-ca.key -CAcreateserial -out exfil-server.crt
```
