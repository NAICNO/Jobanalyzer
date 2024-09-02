# Shared config code for all the scripts

# The experimental slurm data store was created by extracting slurm data on fox by running
#   sacctd -span 2024-03-01,2024-06-01 > the-data.csv
# and then importing it on naic-monitor into a local scatch store using
#   sonalyze add -slurm-sacct -data-dir $SLURMDATADIR < the-data.csv

SLURMDATADIR=~/fox-experiment/data
SONALYZE_SACCT="../../code/sonalyze/sonalyze sacct -data-dir $SLURMDATADIR"
SONALYZE=../../code/sonalyze/sonalyze
SONARDATADIR=~/sonar/data/fox.educloud.no
CONFIGFILE=~/sonar/scripts/fox.educloud.no/fox.educloud.no-config.json
HEATMAP=../../code/heatmap/heatmap
FIRSTDAY=2024-03-01
LASTDAY=2024-05-31
