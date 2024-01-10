# Sonar on the compute nodes

We run `sonar` by means of cron on each node in each cluster.  This samples the node every few
minutes and exfiltrates the sample data to the analysis host by means of the `exfiltrate` program.

A companion program `sysinfo` can be used to probe the system configuration interactively.

On each cluster there is a *sonar home directory*, which is typically on a shared disk.  This
directory holds programs and scripts:

* `sonar` is an executable that samples the system and prints sample data on stdout
* `exfiltrate` is an executable that reads sample data from stdin and transports it over the network to
   the analysis host
* `sysinfo` is an executable that attempts to extract information about a node's configuration
* `sonar.sh` is the script that runs `sonar` and `exfiltrate` with appropriate options
* `sonar-runner.cron` is a `cron` script appropriate for the host to schedule invocations of `sonar.sh`

There is also another file:

* `exfil-auth.txt` contains identity information used by `exfiltrate` in its communication with the
  server.  For more information about this file, see the documentation in `../sonar-server`.

The contents of `sonar.sh` and `sonar-runner.cron` may vary from cluster to cluster, and the
binaries are naturally architecture-dependent.

## ML nodes (cluster name: mlx.hpc.uio.no)

The sonar home directory is `~larstha/sonar` though this may change.  The cron job is run as
`larstha` by means of `crontab`.

`exfil-auth.txt` is stored in `~larstha/.ssh`.


## Fox (cluster name: fox.educloud.no)

The sonar home directory is `/cluster/var/sonar/bin`, owned by the system user `sonar`.  The cron
job is run as the user `sonar` by copying `sonar-runner.cron` to /etc/cron.d/sonar.

`exfil-auth.txt` is stored in the sonar home directory with restrictive ownership and permissions
(not precisely ideal but there you have it).

### Compute/gpu nodes

The following additional conditions have to be met on compute and gpu nodes

* there shall be a user `sonar` with UID 502 (the same UID on all nodes) and homedir `/var/run/sonar`
* user can be nologin
* the directory `/var/run/sonar` shall exist
* the owner of `/var/run/sonar` shall be `sonar.sonar`
* in `/etc/security/access.conf` there shall be the permission `+ : sonar : cron` before the deny-all line
* the owner of `/etc/cron.d/sonar` shall be `root.root`
* cron errors show up in `/var/log/cron`

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
( cd Jobanalyzer/exfiltrate ; go build )
( cd Jobanalyzer/sysinfo ; go build )
```

The executables are in `sonar/target/release/sonar`, `Jobanalyzer/exfiltrate/exfiltrate`, and
`Jobanalyzer/sysinfo/sysinfo` respectively.


## The `sysinfo` utility

In each cluster directory there is a file called `<clustername>-config.json` which holds information
about the configuration of every node in the system.  When nodes are added, removed, or changed, the
`sysinfo` program can be run to extract node information to be inserted into that file.

(Eventually, `sysinfo` may be run automatically on a daily basis.)

For a node without a GPU, simply run `sysinfo`.  If the node is known to have an NVIDIA gpu,
add the `-nvidia` flag.  If it is known to have an AMD gpu, add the `-amd` flag.
