#!/bin/bash

set -e

if [ $# -lt 3 ]; then
	echo "fuck you"
	exit 1
fi

wrkr_env_file=$1
db_env_file=$2
key_file=$3

OAR_JOBID=12345 # test
mapfile -t vms < "/tmp/deploynodes.${OAR_JOBID}"
echo "${vms[0]}" > redis.${OAR_JOBID}
echo "${vms[1]}" > gateway.${OAR_JOBID}
echo "${vms[@]:2}" > workers.${OAR_JOBID}

kadeploy3 -f redis.${OAR_JOBID} --env-file ${db_env_file} -k ${key_file}
kadeploy3 -f gateway.${OAR_JOBID} --env-file ${node_env_file} -k ${key_file}
kadeploy3 -f workers.${OAR_JOBID} --env-file ${node_env_file} -k ${key_file}

cp /tmp/deploynodes.${OAR_JOBID} .