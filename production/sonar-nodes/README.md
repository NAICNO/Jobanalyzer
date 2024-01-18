# Running Sonar on compute nodes

We run `sonar` by means of cron on each node in each cluster: A script samples the node every few
minutes using `sonar` and exfiltrates the sample data to the analysis host by means of the
`exfiltrate` program.

A companion program `sysinfo` can be used to probe the system configuration of the node
interactively.

On each cluster there is a *sonar home directory*, which may or may not be on a shared disk.  This
directory holds these programs and scripts:

* `sonar` is an executable that samples the system and prints sample data on stdout
* `exfiltrate` is an executable that reads sample data from stdin and transports it over the network to
   the analysis host
* `sysinfo` is an executable that attempts to extract information about a node's configuration
* `sonar.sh` is the script that runs `sonar` and `exfiltrate` with appropriate options
* `sonar-runner.cron` is a `cron` script appropriate for the host to schedule invocations of `sonar.sh`

There is also a directory with two files:

* `secrets/exfil-auth.txt` contains identity information used by `exfiltrate` in its communication with the
  server.  For more information about this file, see the documentation in `../sonar-server`.
* `secrets/exfil-ca.crt` is the certificate for the NAIC Certificate Authority, this is used for HTTP upload.

The contents of `sonar.sh` and `sonar-runner.cron` may vary from cluster to cluster, and the
binaries are naturally architecture-dependent.

## ML nodes (cluster name: mlx.hpc.uio.no)

Sonar runs as the user `sonar-runner` and the sonar home directory is `~sonar-runner`.

The cron job is run as `sonar-runner` by copying `sonar-runner.cron` to `/etc/cron.d/sonar`.

Currently, `~sonar-runner` is `/var/lib/sonar-runner` on all nodes.  As this directory is private to
each machine, the binaries and scripts have to be updated everywhere every time there is an update.
This is annoying but there are few machines in this cluster and updates are rare, so it's acceptable.

`secrets` and its files have maximal restrictions (only readable and only by owner).

### A script for installing or updating sonar on the ML nodes

First build sonar and jobanalyzer binaries somewhere (as described below) and create
`sonar-mlnodes.tar.gz` that contains the files in the list above except `secrets/exfil-auth.txt`.

There are no secrets in this tar file, the certificate is freely copyable.  Of course the contents
are sensitive in the sense that they can be corrupted, so it can be copied freely but should not be
left open to write access.

The tar file does not normally have to be created more than once, all the ML nodes have the same
architecture.

On each ML host, do as follows.

Put the tar file somewhere accessible to the sonar-runner user, `/tmp` might work:

```
  $ cp sonar-mlnodes.tar.gz /tmp
```

Create the user if necessary:
```
  $ sudo -i
  # useradd -r -m -d /var/lib/sonar-runner -s /sbin/nologin -c "Sonar monitor account" sonar-runner
```

Then set everything up:
```
  # sudo -u sonar-runner /bin/bash
  $ cd
  $ cp /tmp/sonar-mlnodes.tar.gz .
  $ tar xzf sonar-mlnodes.tar.gz
  $ touch secrets/exfil-auth.txt
  $ chmod go-rwx secrets/exfil-auth.txt
  $ <edit secrets/exfil-auth.txt to add credential>
  $ chmod u-w secrets/exfil-auth.txt
  $ ^D
  # <edit /etc/security/access.conf and add "+ : sonar-runner : cron" before any deny-all lines>
  # cp /var/lib/sonar-runner/sonar-runner.cron /etc/cron.d/sonar
  # ^D
  $ rm /tmp/sonar-mlnodes.tar.gz
```

If you want to test the setup without sending output directly to the production system, then edit
`sonar.sh` to use the address of a test system that has an ingestion system set up.

## Fox (cluster name: fox.educloud.no)

Sonar runs as the user `sonar` and the sonar work directory is `/cluster/var/sonar/bin`, owned by
`sonar`.  The cron job is run as the user `sonar` by copying `sonar-runner.cron` to
`/etc/cron.d/sonar`.

**NOTE** On interactive and login nodes, use the files `sonar-no-slurm.sh` and
`sonar-runner-no-slurm.cron` instead of `sonar.sh` and `sonar-runner.cron`.  The reason for this is
that job numbers must be synthesized by sonar on such nodes.

`secrets/` is stored in the sonar work directory with restrictive ownership and permissions.

### Compute/gpu nodes

The following additional conditions have to be met on compute and gpu nodes

* there shall be a user `sonar` with UID 502 (the same UID on all nodes) and homedir `/var/run/sonar`
* user can be nologin
* the directory `/var/run/sonar` shall exist
* the owner of `/var/run/sonar` shall be `sonar.sonar`
* in `/etc/security/access.conf` there shall be the permission `+ : sonar : cron` before the deny-all line
* the owner of `/etc/cron.d/sonar` shall be `root.root`
* cron errors show up in `/var/log/cron`

For new and reinstalled nodes to just work:

* the access.conf setting shall be reflected in `master:/install/dists/fox/syncfiles-el9/etc/security/access.conf`
* the script `master:/install/postscripts/fox_sonar` shall contain commands to set up home directory and permissions

### Interactive nodes

The following additional conditions have to be met on interactive nodes and login nodes

* there shall be a user `sonar` with UID 502 (the same UID on all nodes) and homedir `/home/sonar`
* user can be nologin
* the directory `/home/sonar` shall exist
* the owner of `/home/sonar` shall be `sonar.whatever`
* normally `/etc/security/access.conf` allows everything, otherwise see above
* the owner of `/etc/cron.d/sonar` shall be `root.root`
* cron errors show up in the journal

## Building `sonar`, `exfiltrate` and `sysinfo`

You need to install or load compilers for Go 1.20 or later (for `exfiltrate`) and Rust 1.65 or later
(for `sonar`).  Then:

```
git clone https://github.com/NordicHPC/sonar.git
( cd sonar ; cargo build --release )

git clone https://github.com/NAICNO/Jobanalyzer.git
( cd Jobanalyzer/code/exfiltrate ; go build )
( cd Jobanalyzer/code/sysinfo ; go build )
```

The executables are in `sonar/target/release/sonar`, `Jobanalyzer/code/exfiltrate/exfiltrate`, and
`Jobanalyzer/code/sysinfo/sysinfo` respectively.

## The `sysinfo` utility

In each cluster directory there is a file called `<clustername>-config.json` which holds information
about the configuration of every node in the system.  When nodes are added, removed, or changed, the
`sysinfo` program can be run to extract node information to be inserted into that file.

(Eventually, `sysinfo` may be run automatically on a daily basis.)

For a node without a GPU, simply run `sysinfo`.  If the node is known to have an NVIDIA gpu,
add the `-nvidia` flag.  If it is known to have an AMD gpu, add the `-amd` flag.

## Updating a live server

One does not simply copy new executables into place as this will create all sorts of problems.

Generally the safest bet is to spin down the cron job by removing `/etc/cron.d/sonar` or by
inserting an `exit 0` line at the beginning of sonar.sh.  (On the ML nodes this has to be done once
per node, on Fox the script is shared.)

Then, once sonar and exfiltrate are not running (`pgrep exfil` returning no hits is a good
indication), the new executables can be put in place, and the cron job restarted.

Alternatively, it may be OK to to generate new executables, copy them to `~/sonar` under different
names, and then `mv` them into place, as this should be atomic and any references to the existing
executables will keep those alive behind the scenes.  But this is not well tested.
