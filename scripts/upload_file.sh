#!/bin/bash
# upload file to host

set -e

if [ $# -lt 2 ]; then
	echo "fuck you"
	exit 1
fi

fichier=$1
rhost=$2

if ! nc -z $rhost 22; then
	echo "remote host $rhost not open in port 22"
	exit 2
fi

opts="-o StrictHostKeyChecking=no"

if [ -d "$fichier" ]; then
	opts="$opts -r"
fi

scp $opts $fichier root@$rhost:~/

