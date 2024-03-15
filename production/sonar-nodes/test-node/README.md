These are scripts that can be run on any linux system and will upload sonar data from the local
system to naic-monitor.uio.no.  I use them to test changes in the logic for running sonar and
exfiltrating data.

To set up a local test "node", create ~/sonar-test-node and populate it as follows:

- symlinks to the *.sh files in this directory
- a sonar executable
- a subdirectory `secrets`
- in that directory, a file `upload-auth.netrc` with the following contents, where the string PASSWORD
  is the password for this cluster as recorded in the upload password file on the server:
  ```
  machine naic-monitor.uio.no login naic-monitor.uio.no password PASSWORD
  ```

Now, running eg `./sonar-batchless.sh` in that directory will run sonar, capture the output, and
(after a suitable delay) exfiltrate it to `naic-monitor.uio.no`, where it will appear in the data
store for the `naic-monitor.uio.no` cluster.

(One could argue that the test cluster should have a more obvious name, eg, `test-cluster`, but it
is what it is.)
