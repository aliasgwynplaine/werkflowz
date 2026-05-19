#!/bin/bash

set -e

fuckyou () { echo "fuck you"; }

# sleeps printing dots...
dot_sleep () {
	if [ $# -lt 1 ]; then
		fuckyou
		return
	fi

	for i in $(seq $1); do
		sleep 1
		echo -n "."
	done
}

if [ $# -lt 4 ]; then
	echo "usage: $0 <nb-nodes> <h:mm:ss> <node-env-file> <db-env-file> <key-file>"
	exit 1
fi

nb_nodes=$1
walltime=$2
node_env_file=$(realpath $3)
db_env_file=$(realpath $4)
key_file=$(realpath $5)

if [ $nb_nodes -lt 3 ]; then
	echo "Cannot deploy the experiment with less than 3 nodes"
	exit 3
fi

echo "KEY_FILE -> $key_file"
echo "NODE_ENV_FILE -> $node_env_file"
echo "DB_ENV_FILE -> $db_env_file"

tmpfile=$(mktemp "./lx2k2XXXXXX")

# deploy the VMs
oarsub -t deploy -l /nodes=$nb_nodes,walltime=$walltime \
"bash arch_script.sh ${node_env_file} ${db_env_file} ${key_file} && while true; do sleep 1; done" | tee $tmpfile

OAR_JOBID=$(grep OAR_JOB_ID $tmpfile | cut -d "=" -f 2)
rm $tmpfile

echo "Waiting for nodes..."

while [ ! -f "deploynodes.${OAR_JOBID}" ]; do
	dot_sleep 37
	#echo "looking for deploynodes.${OAR_JOBID}"
done

# recover VMs info
mapfile -t vms < "deploynodes.${OAR_JOBID}"

echo "Hostnames: ${vms[@]}"
echo "Waiting for hostnames to boot..."
dot_sleep 43

# wait for ssh	
while [[ $gtfo -lt 5 ]]; do
	echo "attempting to connect with $vms..."
	gtfo=0
	declare -A ready

	for vm in ${vms[@]}; do
		echo -n "Trying $vm... "

		if nc -z $vm 22; then
			ready[$vm]=1
			echo "ok!"
		else
			ready[$vm]=0
			echo "error"
		fi
	done

	for vm in ${vms[@]}; do
		gtfo=$(( gtfo + ${ready[$vm]} ))
	done

	if [[ $gtfo -lt 5 ]]; then
		echo "servers are not ready."
		echo "Waiting..."
		dot_sleep 37
	else
		echo "servers ready"
		continue
	fi

	echo "Retrying..."
	#vms=$(oarstat -u -f -j ${OAR_JOBID} | grep assigned_hostnames | cut -d "=" -f 2)
	#vms="${vms#[[:space:]]}"
	#vms="${vms%[[:space:]]}"
	#read -r -a vms <<< "$(echo $vms | sed s/\+/\ /g)"
done

echo "nodes are ready!"
# todo
