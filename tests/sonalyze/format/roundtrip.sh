output1=$($SONALYZE parse --from 2023-10-04 --fmt=csvnamed,roundtrip,nodefaults -- roundtrip.csv)
CHECK format_csv_roundtrip1 \
      'v=0.7.0,time=2023-10-04T07:40Z,host=ml4.hpc.uio.no,cores=64,user=einarvid,job=1269178,cmd=python3,cpu%=1714.2,cpukib=273864972,cputime_sec=10192,rolledup=69' \
      "$output1"

TEMPDIR=roundtrip_tmpdir
TEMPFILE=$TEMPDIR/out$$
rm -rf $TEMPDIR
mkdir -p $TEMPDIR
echo $output1 > $TEMPFILE
output2=$($SONALYZE parse --from 2023-10-04 --fmt=csvnamed,roundtrip,nodefaults -- $TEMPFILE)
CHECK format_csv_roundtrip2 "$output" "$output2"
rm -rf $TEMPDIR
