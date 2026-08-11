#!/bin/bash
# upload file to host

set -e

if [ $# -lt 2 ]; then
	echo "fuck you"
	exit 1
fi

# if you invert the arguments you will have a more flexible script
fichier=$1
hostlist=$2

mapfile -t rhosts < "$hostlist"

#opts="-o StrictHostKeyChecking=no"
opts=(-avz -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null")

#if [ -d "$fichier" ]; then
#	opts="-r $opts"
#fi

for rhost in ${rhosts[@]}; do
	if ! nc -z $rhost 22; then
		echo "remote host $rhost not open in port 22"
		exit 2
	fi
	echo "[*] coping files to $rhost"
	rsync ${opts} $fichier root@$rhost:~/
	echo "[*] $rhost done!"
done

