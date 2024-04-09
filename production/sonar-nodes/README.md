# Running Sonar on nodes in clusters

We run `sonar` by means of cron on each node in each cluster.  Cron runs two scripts:

* One script samples the node every few minutes using `sonar ps` and then exfiltrates the sample
  data to the analysis host by means of `curl`.
* Another script runs `sonar sysinfo` daily and on reboot to probe the system configuration of the
  node, and these data are also exfiltrated with `curl`.

On each cluster there is a *sonar home directory*, ideally on a shared disk.  This directory holds
the following programs and scripts.

* `sonar` is an executable that samples the system and prints sample data on stdout.
* `sonar-<config>.sh` is the script that runs `sonar ps` to sample process data, here `<config>` is
  `slurm` for nodes controlled by slurm and `batchless` for nodes without any batch system at all
  (typically login, interactive, and ad-hoc nodes).
* `sysinfo.sh` is the ditto script that runs `sonar sysinfo` to probe the system
  configuration.
* `sonar-config.sh` (literally the name of it) is a shared configuration file that is read by the
  two scripts.
* `sonar-runner-<config>.cron` is a `cron` script appropriate for a node to schedule invocations of
  `sonar-<config>.sh` and `sysinfo.sh`, it's the master copy for the crontab that has been
  copied to `/etc/cron.d/sonar` on each node.  If the cluster has both batch nodes and batchless
  nodes then there will be two crontabs here, one for each node type.  The correct file must be
  copied to `/etc/cron.d/sonar` on each node.

In the sonar home directory there is also a subdirectory `secrets/` with authentication information:

* `secrets/upload-auth.netrc` contains identity information used by `curl` in its communication with
  the server.  For more information about this file, see the documentation in
  `../jobanalyzer-server`.  The format is a `netrc` file:
  ```
  machine naic-monitor.uio.no login LOGIN password PASSWORD
  ```
  where `LOGIN` and `PASSWORD` must be obtained from the admin of `naic-monitor.uio.no` and will
  be different for different clusters.

The `secrets/` subdirectory should have mode 0700 or 0500 and the files within it should have mode
0600 or 0400.

The contents of all the scripts and the crontab may vary a little from cluster to cluster (see
below), and the binaries are naturally architecture-dependent.


## ML nodes (cluster name: mlx.hpc.uio.no)

Sonar runs as the user `sonar-runner` and the sonar home directory is `~sonar-runner`, owned by
`sonar-runner:sonar-runner`.  The scripts and crontabs are set up to use `sonar_dir=$HOME`.

The cron job is run as `sonar-runner` by copying `sonar-runner-batchless.cron` to `/etc/cron.d/sonar`.

**NOTE**: You may wish to change the `MAILTO` definition in the cron script first.

Since none of the nodes have slurm, the sonar script is `sonar-batchless.sh`.

Currently, `~sonar-runner` is `/var/lib/sonar-runner` on all nodes.  As this directory is private to
each machine, the binaries and scripts have to be updated everywhere every time there is an update.
It is hard to use a shared directory since the UID of `sonar-runner` is not the same on every
machine.

### A procedure for installing or updating sonar on the ML nodes

Let's assume you have a checkout of the repos for Sonar and Jobanalyzer (see later for instructions)
and you are in the parent directory of the two repos.  We'll construct a temporary directory
`upload-tmp` and populate it with the necessary directory structure and files.

```
mkdir upload-tmp
( cd sonar ; cargo build --release )
cp sonar/target/release/sonar upload-tmp
cp Jobanalyzer/production/sonar-nodes/mlx.hpc.uio.no/*.{sh,cron} upload-tmp
cd upload-tmp
mkdir secrets
echo "machine naic-monitor.uio.no login mlx.hpc.uio.no password PASSWORD" > secrets/upload-auth.netrc
chmod -R go-rwx secrets
```

Now, consider whether to edit these files:

- `*-batchless.sh` and `sonar-runner-batchless.cron`, to change the sonar root directory
- `sonar-config.sh`, to change any other settings
- `sonar-runner-batchless.cron`, to change cron's `MAILTO` variable

Now we can create a tar file that contains everything except the password (and is therefore freely
copyable):

```
tar czf sonar-mlnodes.tar.gz sonar *.sh *.cron secrets
```

The tar file does not normally have to be created more than once, all the ML nodes have the same
architecture.

On each ML node, do as follows.

Put the tar file somewhere accessible to the sonar-runner user, `/tmp` might work if you don't want
to open your build directory to the world:

```
  $ cp sonar-mlnodes.tar.gz /tmp
```

Create the user if necessary:
```
  $ sudo -i
  # useradd -r -m -d /var/lib/sonar-runner -s /sbin/nologin -c "Sonar monitor account" sonar-runner
  # <edit /etc/security/access.conf and add "+ : sonar-runner : cron" before any deny-all lines>
```

Then set everything up:
```
  # sudo -u sonar-runner /bin/bash
  $ cd
  $ tar xzf /tmp/sonar-mlnodes.tar.gz
  $ <edit secrets/upload-auth.netrc to add the PASSWORD>
  $ ^D
  # cp /var/lib/sonar-runner/sonar-runner-batchless.cron /etc/cron.d/sonar
  # ^D
  $ rm /tmp/sonar-mlnodes.tar.gz
```

If you want to test the setup without sending output directly to the production system, then edit
`sonar-config.sh` to use the address of a test system that has data ingestion set up.


## Fox (cluster name: fox.educloud.no)

Sonar runs as the user `sonar` and the sonar work directory is `/cluster/var/sonar`, owned by
`sonar:sonar`.  The cron job is run as the user `sonar` by copying `sonar-runner-<config>.cron` to
`/etc/cron.d/sonar`, for the appropriate `<config>` for the node.

**NOTE**: You may wish to change the `MAILTO` definition in the cron script first.

On Fox we use a lock file to prevent multiple instances of Sonar to run at the same time, the lock
file directory is currently `/var/tmp`.


### Compute/gpu nodes

On these nodes, the cron script is `sonar-runner-slurm.cron` and it will run the scripts
`sonar-slurm.sh` and `sysinfo.sh`.

The following additional conditions have to be met on compute and gpu nodes

* there shall be a user `sonar` with UID 502 (the same UID on all nodes) and homedir `/var/lib/sonar`
* user can be nologin
* the directory `/var/lib/sonar` shall exist
* the owner of `/var/lib/sonar` shall be `sonar.sonar`
* in `/etc/security/access.conf` there shall be the permission `+ : sonar : cron` before the deny-all line
* the owner of `/etc/cron.d/sonar` shall be `root.root`
* cron errors show up in `/var/log/cron`

To make sure sonar is set up on new and reinstalled nodes:

* the access.conf setting shall be reflected in `master:/install/dists/fox/syncfiles-el9/etc/security/access.conf`
* the script `master:/install/postscripts/fox_sonar` shall contain commands to set up home directory and permissions


### Interactive/login nodes

On these nodes, the cron script is `sonar-runner-batchless.cron` and it will run the scripts
`sonar-batchless.sh` and `sysinfo.sh`.  The reason for this difference from the compute
nodes is that job numbers must be synthesized by Sonar.

The following additional conditions have to be met on interactive nodes and login nodes

* there shall be a user `sonar` with UID 502 (the same UID on all nodes) and homedir `/home/sonar`
* user can be nologin
* the directory `/home/sonar` shall exist
* the owner of `/home/sonar` shall be `sonar.whatever`
* normally `/etc/security/access.conf` allows everything, otherwise see above
* the owner of `/etc/cron.d/sonar` shall be `root.root`
* cron errors show up in the journal

### Setting up

The setup procedure (building, creating a tar file, installing it) is largely as for the ML nodes
except that the user is `sonar` and not `sonar-runner`.


## Saga (cluster name: saga.sigma2.no)

When complete, the setup will be as for Fox (see above) except the work directory is
`/cluster/shared/sonar` (instead of `/cluster/var/sonar`), and there is currently no requirement to
use a lock file.


## Test cluster (cluster name: naic-monitor.uio.no)

For testing, the production server will receive Sonar and Sysinfo data for the cluster
`naic-monitor.uio.no` with arbitrary host names and at arbitrary times.  See
[test-node/README.md](test-node/README.md) for more information about how to use this.


## Adding new compute clusters

Information about how to set up sonar on the server is in
[../jobanalyzer-server/README.md](../jobanalyzer-server/README.md).

(Possibly stale information) To complement information above, see [the PR that added everything for
Saga](https://github.com/NAICNO/Jobanalyzer/pull/364) for an example of what a new node
configuration might look like.


## Building `sonar`

You need to install or load a compiler for Rust 1.74 or later to build `sonar`.

```
git clone https://github.com/NordicHPC/sonar.git
cd sonar
cargo build --release
```

The executable will appear in `target/release/sonar`.  Test it by running `target/release/sonar ps`.

To check out Jobanalyzer:
```
git clone https://github.com/NAICNO/Jobanalyzer.git
```

## Updating a live server

One does not simply copy new executables and scripts into place as this will create all sorts of
problems.

If the only thing that has changed is the `sonar` executable then it can normally be `mv`'d into
place without stopping anything; any running sonar processes will have a handle to the original file
while still executing.

For everything else, generally the safest bet is to spin down the cron job by removing
`/etc/cron.d/sonar`.  This has to be done once per node.

Then, once `sonar` is not running (`pgrep sonar` returning no hits is a good indication), the new
files can be put in place, and the cron job restarted.
