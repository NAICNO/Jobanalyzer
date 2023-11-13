# How to set up Jobanalyzer and Sonar on Fox

We will set up a cron job to run sonar on every node and another cron job to run analysis
asynchronously on a non-node.

## Common preparation for Sonar and Jobanalyzer

You need an `ec-whatever` user account on Fox to hold shared files and to run the analysis jobs (for now).

Log in as `ec-whatever` on a login node.  Then:

```
    git clone https://github.com/NAICNO/Jobanalyzer.git
    cd Jobanalyzer
    module load Rust/1.65.0-GCCcore-12.2.0
    module load Go/1.20.4
    ./build.sh
```

(For Go, anything after 1.20 should be fine.  For Rust, anything after 1.65 should be fine, and I
don't think the GCCCore version matters.)

As root on the login node or some other node with access to the file systems:

```
    mkdir /cluster/var/sonar
    cd /cluster/var/sonar
    mkdir bin data
    cp ~ec-whatever/Jobanalyzer/sonar/target/release/sonar bin
    cp ~ec-whatever/Jobanalyzer/production/fox/sonar.sh bin
    chown -R sonar.sonar .
    cp ~ec-whatever/Jobanalyzer/production/fox/jobanalyzer.cron bin
```

Note above that the owner of the cron file remains root.root, this is important.

The sonar cron job can now be changed everywhere by editing (as root)
`/cluster/var/sonar/bin/jobanalyzer.cron`, and notably the sonar job can be stopped everywhere by
commenting out all the lines in that file.


## Setting up sonar

As you do the following, if you have the fox dashboard running and everything is going ok, new nodes
should appear there automatically as they come on line.

### Per node (non-int nodes)

Now **as root** on each non-int node where you want to run sonar:

Check that the user exists:

```
    grep sonar /etc/passwd
```

You should see (note the uid must be the same everywhere):

```
    sonar:x:502:502::/var/run/sonar:/sbin/nologin
```

If the home directory does not exist, create it:

```
    mkdir /var/run/sonar
    chown sonar.sonar /var/run/sonar
```

Then enable cron permissions for sonar:

```
    vi /etc/security/access.conf
    # Before the last line that reads "- : ALL : ALL" add a line that says "+ : sonar : cron"
```

Then extract information about the node, on JSON form.  It may be easiest to run the `sysinfo`
program in the Jobanalyzer repo.  Assuming you have compiled everything with `build.sh` above and
the binaries are still present, then:

```
    ~ec-whatever/Jobanalyzer/sysinfo/sysinfo
```

will print a JSON object that can be pasted into a JSON file; save it for later use (see below).

If the node has a GPU, then pass the `-nvidia` or `-amd` switches to `sysinfo` as appropriate to
capture that information as well.

Finally, start the cron job:

```
    ln -s /cluster/var/sonar/bin/jobanalyzer.cron /etc/cron.d/sonar
```

Any CRON errors will show up in /var/log/cron, watch this file for a bit.


### Per node (int nodes)

Now **as root** on each int node where you want to run sonar.

```
$ grep sonar /etc/passwd
```

You should see
```
sonar:x:502:502::/home/sonar:/bin/bash
```

Then check the home directory:

```
$ ls -ld /home/sonar
drwx------ 2 sonar 1000 90 Nov  8 13:41 /home/sonar
```

The group doesn't look right but oh well, seems to work... If the home directory isn't there, you
need to make one, see the non-int instructions above.

There's no need to edit `access.conf` for the int-nodes, as the current file allows everyone to do
everything.

Finally, start the cron job:

```
    ln -s /cluster/var/sonar/bin/jobanalyzer.cron /etc/cron.d/sonar
```

TODO: Cron errors show up where???


## Setting up analysis

First, create a directory and copy files into it:

```
    mkdir ~/sonar
    cp ~/Jobanalyzer/production/fox/*.sh ~/sonar
    cp ~/Jobanalyzer/production/fox/fox.json ~/sonar
    cp ~/Jobanalyzer/production/fox/jobanalyzer-analysis.cron ~/sonar
    cp ~/Jobanalyzer/sonalyze/target/release/sonalyze ~/sonar
    cp ~/Jobanalyzer/naicreport/naicreport ~/sonar
    cp ~/Jobanalyzer/loginfo/loginfo ~/sonar
    ln -s /cluster/var/sonar/data ~/sonar/data
    ln -s ~/.ssh/ubuntu-vm.pem ~/sonar
```

First, edit `fox.json` and add information about new nodes you obtained earlier.  Watch the syntax!
In particular, no commas before closing brackets, but commas between sibling items.

Second, you need to add the file ~/.ssh/ubuntu-vm.pem, which is the key used for uploading.

Then start the cron job:
```
    crontab ~/sonar/jobanalyzer-analysis.cron
```
