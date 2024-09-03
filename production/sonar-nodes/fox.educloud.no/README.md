Mostly this is as documented in the Jobanalyzer repo (https://github.com/NAICNO/Jobanalyzer).  With these
additions:

- there is a cron job on login-1 as the user `sonar`, it runs sacctd.sh once an hour, see sacctd-runner.cron
