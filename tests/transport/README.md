# Things to test

This list is very long indeed.  Here's a start:

## Functionality

- Non-zero sending window
- Multiple cluster names
- More elaborate data
- Test exit codes more generally, should not allow failures

## Error cases for infiltrate

- data-path wrong (does not exist, wrong permissions), retry
- port bad
- auth file bad or not found
- port taken

## Error cases for exfiltrate

- server not listening / retry
- error return from server
- missing args
- source format wrong
- output format wrong
- target address not valid URL (syntax)
- target address not valid URL (unsupported scheme)
- target address not found
- bad authentication: user/pass mismatch
- bad authentication: infiltrate expects auth but exfiltrate does not send
- bad authentication: infiltrate expects no auth but exfiltrate sends
- bad authentication: auth file not found or wrong syntax
