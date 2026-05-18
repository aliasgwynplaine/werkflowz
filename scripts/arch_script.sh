#!/bin/bash

set -e

if [ $# -lt 3 ]; then
	echo "fuck you"
	exit 1
fi

node_env_file=$1
db_env_file=$2
key_file=$3

mapfile -t vms < "/tmp/deploynodes.${OAR_JOBID}"
echo "${vms[0]}" > redis.${OAR_JOBID}
echo "${vms[1]}" > gateway.${OAR_JOBID}

for wrkr in "${vms[@]:2}"; do
	echo "$wrkr" >> workers.${OAR_JOBID}
done

cp /tmp/deploynodes.${OAR_JOBID} .

kadeploy3 -f redis.${OAR_JOBID} --env-file ${db_env_file} -k ${key_file} &
kadeploy3 -f gateway.${OAR_JOBID} --env-file ${node_env_file} -k ${key_file} &
kadeploy3 -f workers.${OAR_JOBID} --env-file ${node_env_file} -k ${key_file} &


