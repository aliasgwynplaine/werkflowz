#!/bin/bash
# command & control - c2

set -e

usage () {
	echo "usage: $0 <command> <opts>"
	echo
	echo "commands:"
	echo "	deploy-nodes <nb> <h:mm:ss> <node-env-file> <db-env-file> <key-file>"
	echo "	upload-server <nightcore.host>"
	echo "	upload-app <app> <nighcore.host>"
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

	upload-server)
		echo "upload-server!"

		if [ $# -ne 2 ]; then
			usage
			exit 2
		fi
	;;

	upload-app)
		echo "upload-app!"

		if [ $# -ne 2 ]; then
			usage
			exit 2
		fi
	;;

	run-gateway)
		echo "run-gateway!"

		if [ $# -ne 2 ]; then
			usage
			exit 2
		fi
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

