output=$($SLURMINFO -background auxfile.json -input slurminfo-test.txt)
CHECK slurminfo_smoke "$(cat slurminfo-expect.txt)" "$output"
