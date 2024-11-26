output=$($SONALYZE parse --fmt=csvnamed,all,nodefaults -- nodefault.csv)
CHECK format_csv_nodefault \
      'version=0.7.0,localtime=2023-10-04 07:40,host=ml4.hpc.uio.no,cores=64,user=einarvid,job=1269178,cmd=python3,cpu_pct=1714.2,mem_gb=261,cputime_sec=10192,rolledup=69' \
      "$output"

output=$($SONALYZE parse --fmt=json,all,nodefaults -- nodefault.csv)
CHECK format_csv_nodefault \
      '[{"version":"0.7.0","localtime":"2023-10-04 07:40","host":"ml4.hpc.uio.no","cores":"64","user":"einarvid","job":"1269178","cmd":"python3","cpu_pct":"1714.2","mem_gb":"261","cputime_sec":"10192","rolledup":"69"}]' \
      "$output"
