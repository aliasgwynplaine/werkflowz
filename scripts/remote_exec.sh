#!/bin/bash

if [ $# -lt 2 ]; then
	echo "usage: $0 <host> <script.sh> <args...>"
	exit 1
fi

rhost=$1
myscript=$2

# je m'en fous de la securite :P
ssh -o StrictHostKeyChecking=no root@$rhost 'bash -s ' < $myscript ${@:3} &
