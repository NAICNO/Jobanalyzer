output=$($SONALYZE jobs -u nissen --min-samples=1 -f 2023-10-04 --fmt=csv,job,user -- select_user.csv)
CHECK one_user_1 "1269178,nissen" "$output"

output=$($SONALYZE jobs -u zappa --min-samples=1 -f 2023-10-04 --fmt=csv,job,user -- select_user.csv)
CHECK one_user_2 "1269179,zappa" "$output"

# `--exclude-user` implies `-u-`.
output=$($SONALYZE jobs --exclude-user satan --min-samples=1 -f 2023-10-04 --fmt=csv,job,user -- select_user.csv | sort)
CHECK two_users "1269178,nissen
1269179,zappa" "$output"

# LOGNAME selects the user in the absence of -u
output=$(LOGNAME=nissen $SONALYZE jobs --min-samples=1 -f 2023-10-04 --fmt=csv,job,user -- select_user.csv)
CHECK user_from_logname "1269178,nissen" "$output"

# USER selects the user in the absence of -u and LOGNAME
output=$(LOGNAME="" USER=nissen $SONALYZE jobs --min-samples=1 -f 2023-10-04 --fmt=csv,job,user -- select_user.csv)
CHECK user_from_user "1269178,nissen" "$output"

# --user - selects all
output=$($SONALYZE jobs --user - --min-samples=1 -f 2023-10-04 --fmt=csv,job,user -- select_user.csv | sort)
CHECK all_users "1269177,satan
1269178,nissen
1269179,zappa" "$output"

