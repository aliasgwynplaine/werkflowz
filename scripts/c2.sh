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
	echo "	run-gateway <gateway.host> <experiment-name>"
	echo "	kill-gateway <gateway.host>"
	echo "	run-engine <workers.host> <experiment-name> <gateway.host>"
	echo "	kill-all-engines <workers.host>"
	echo "	run-launcher <worker> <expriment-name> <func_id> <gateway.host>"
	echo "	kill-all-launchers <workers.host>"
	echo "	redis-setup <redis.host>"
	echo "	ccmesh-redis-run <redis.host>"
	echo "	ccmesh-server-build <workers.host>"
	echo "	ccmesh-server-run <workers.host>"
	echo "	hit <gateway.host> <func_name> <data>"
	echo "	remote-kill <file.host> <procname>"
	echo "	retrieve-results"
	echo "	clean-all"
	echo "	help"
	echo
}

show-experiments() {
	echo "experiments:"
	echo "	incrementor"
	echo "	writernreader"
	echo "	2branch"
	echo "	3branch"
	echo
}

verify-experiment() {
	EXPERIMENT=$1

	case "$EXPERIMENT" in
		incrementor)
			return 0
			;;
		incrementorNotRet)
			return 0
			;;
		writernreader)
			return 0
			;;
		2branch)
			return 0
			;;
		3branch)
			return 0
			;;
		*)
			echo "Unknown experiment: $EXPERIMENT"
			exit 1
			;;
	esac
}

if [ $# -lt 1 ]; then
	usage
	exit 1
fi

COMMAND=$1

case "$COMMAND" in
deploy-nodes)
	if [ $# -ne 6 ]; then
		usage
		exit 5
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
	if [ $# -ne 3 ]; then
		usage
		exit 2
	fi

	if [ ! -f "$2" ]; then
		echo "file not found: $2"
		exit 127
	fi

	verify-experiment $3
	mapfile -t gw < $2
	bash remote_exec.sh $gw run_gateway.sh $3
	sleep 1
	;;

kill-gateway)
	if [ $# -ne 2 ]; then
		usage
		exit 2
	fi

	mapfile -t gw < $2
	ssh root@$gw 'killall gateway'
	;;

run-engine)
	if [ $# -ne 4 ]; then
		usage
		exit 3
	fi

	mapfile -t wrkrs < $2
	experiment=$3
	node_id=0
	mapfile -t gateway_addr < $4 

	for wrkr in ${wrkrs[@]}; do
		bash remote_exec.sh $wrkr run_engine.sh $experiment $node_id $gateway_addr
		node_id=$((node_id + 1))
		sleep 1
	done

	sleep 2
	;;

kill-all-engines)
	if [ $# -ne 2 ]; then
		usage
		exit 2
	fi

	mapfile -t wrkrs < $2

	for wrkr in ${wrkrs[@]}; do
		ssh root@$wrkr 'pgrep engine && killall engine; exit 0'
	done
	;;

run-launcher)
	if [ $# -ne 5 ]; then
		usage
		exit 4
	fi

	mapfile -t gateway_addr < $5 
	bash remote_exec.sh $2 run_launcher.sh $3 $4 $gateway_addr
	sleep 1
	;;

kill-all-launchers)
	if [ $# -ne 2 ]; then
		usage
		exit 2
	fi

	mapfile -t wrkrs < $2

	for wrkr in ${wrkrs[@]}; do
		ssh root@$wrkr 'pgrep launcher && killall launcher; exit 0'
	done
	;;

kill-all)
	if [ $# -ne 2 ]; then
		usage
		exit 2
	fi

	mapfile -t gw < gateway.$2
	mapfile -t wrkrs < workers.$2

	ssh root@$gw 'pgrep gateway && killall gateway; exit 0'

	for wrkr in ${wrkrs[@]}; do
		ssh root@$wrkr 'pgrep launcher && killall launcher; exit 0'
		ssh root@$wrkr 'pgrep engine && killall engine; exit 0'
	done
	;;

redis-setup)
	if [ $# -ne 2 ]; then
		usage
		exit 1
	fi

	mapfile -t redis_server < $2
	bash remote_exec.sh $redis_server redis_setup.sh
	sleep 1
	;;

ccmesh-redis-run)
	if [ $# -ne 2 ]; then
		usage
		exit 1
	fi

	mapfile -t redis_server < $2
	bash remote_exec.sh $redis_server ccmesh_redis_build.sh
	;;

ccmesh-server-build)
	if [ $# -ne 2 ]; then
		usage
		exit 1
	fi

	mapfile -t workers < $2

	for wrkr in $workers; do
		bash remote_exec.sh $wrkr ccmesh_server_build.sh
	done

	sleep 10
	;;

ccmesh-server-run)
	if [ $# -ne 2 ]; then
		usage
		exit 1
	fi

	mapfile -t workers < $2
	idx=0

	for wrkr in ${workers[@]}; do
		bash remote_exec.sh $wrkr ccmesh_server_run.sh $idx
		idx=$(( idx + 1 ))
	done


	echo "orders were sent"
	sleep 10

	;;

hit)
	if [ $# -lt 4 ]; then
		usage
		exit 3
	fi

	mapfile -t gateway_addr < $2
	entry=$3
	payload=$4
	curl -X POST -d "$payload" http://$gateway_addr:8080/function/$entry
	;;

remote-kill)
	if [ $# -lt 3 ]; then
		usage
		exit 2
	fi

	mapfile -t rhost < $2
	ssh root@$rhost "killall $3"
	sleep 1
	;;

clean-all)
	find . -type f -regex './workers\.[0-9]+' -delete
	find . -type f -regex './gateway\.[0-9]+' -delete
	find . -type f -regex './redis\.[0-9]+' -delete
	find . -type f -regex './deploynodes\.[0-9]+' -delete
	find . -type f -regex './OAR\.[0-9]+\.std*' -delete

	echo "All clean"
	;;

help)
	usage
	exit 0
	;;

*)
	echo "unknown command: $COMMAND"
	echo
	usage
	exit 127
	;;

esac

