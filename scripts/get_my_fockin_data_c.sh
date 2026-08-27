#!/bin/bash
set -e

if [ $# -ne 2 ]; then
	echo "usage: $0 <job-id> <micro-bench>"
	exit 1
fi

jobid=$1
mb=$2

case "$mb" in
	micro-ccmesh)
		runner="run_micro_ccmesh_launchers.sh"
		;;
	micro-ccmesg)
		runner="run_micro_ccmesg_launchers.sh"
		;;
	micro-ccmesgbox)
		runner="run_micro_ccmesgbox_launchers.sh"
		;;
	*)
		echo "Unknown micro-bench: $mb"
		exit 1
		;;
esac

mapfile -t redishost < "redis.${jobid}"
mapfile -t schedhost < "scheduler.${jobid}"
mapfile -t workers  < "workers.${jobid}"

rpss="50 100 150 200 225 250 275 300 325 350 400 425 450 475 500 1000 1500 2000 2500 3000 3500 4000"
#rpss="300 350 400 500"

if [ ! -d "data" ]; then
	mkdir data
	mkdir -p data/wrk
	mkdir -p data/mon
fi

echo $rps > data/.metadata
echo $mb >> data/.metadata

for rps in $rpss; do
	echo "[*] Running ${rps} req/s on $runner"
	echo "[*] Flushing all data"
	bash remote_exec_sync.sh $redishost flushall.sh
	wait

	bash $runner $jobid
	wait
	
	bash remote_exec_sync.sh $schedhost run_wrk2.sh $rps
	wait

	scp root@${workers[0]}:~/ccmesh-server/monitor.txt data/mon/${rps}_monitor.txt
done

scp root@$schedhost:~/data/* data/wrk/

echo "Done!"
