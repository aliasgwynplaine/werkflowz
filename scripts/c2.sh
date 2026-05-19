#!/bin/bash
# command & control - c2

set -e

usage () {
	echo "usage: $0 <command> <opts>"
	echo
	echo "commands:"
	echo "	deploy-nodes <nb> <h:mm:ss> <node-env-file> <db-env-file> <key-file>"
	echo "	upload <file> <host>"
	echo "	upload-to-hosts <file> <file.host>"
	echo "	run-gateway <func_config.json>"
	echo "	run-engine <func_config.json>"
	echo "	run-launcher <ip> <func_id> <fprocess>" # just using go this time
	echo "	redis-setup"
	echo "	clean"
	echo "	help"
	echo
	echo "WARNING: the file nightcore.host must exist."
}


if [ $# -lt 1 ]; then
	usage
	exit 2
fi

COMMAND=$1

case "$COMMAND" in
	deploy-nodes)
		if [ $# -ne 6 ]; then
			usage
			exit 2
		fi

		bash deploy_nodes.sh $2 $3 $4 $5 $6
	;;

	upload)
		if [ $# -ne 3 ]; then
			usage
			exit 2
		fi

		bash upload_file.sh $2 $3
	;;

	upload-to-hosts)
		if [ $# -ne 3 ]; then
			usage
			exit 2
		fi

		bash upload_to_hosts.sh $2 $3
	;;

	run-gateway)
		echo "run-gateway!"

		if [ $# -ne 2 ]; then
			usage
			exit 2
		fi

		#r
	;;

	help)
		usage
		exit 0
	;;

	*)
		echo "unknown command: $COMMAND"
		echo
		usage
		exit 3
	;;
esac

