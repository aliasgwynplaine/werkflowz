#!/bin/bash
set -e

if [ $# -ne 1 ]; then
	echo "usage: $0 <job-id>"
	exit 1
fi

jobid=$1

mapfile -t redishost < "redis.${jobid}"
mapfile -t schedhost < "scheduler.${jobid}"
mapfile -t workers  < "workers.${jobid}"

rpss="50 100 150 200 225 250 275 300 325 350 400 425 450 475 500 1000 1500 2000 2500 3000 3500 4000"
#rpss="300 350 400 500"

if [ ! -d "data" ]; then
	mkdir data
fi

for rps in $rpss; do
	echo "[*] Running ${rps} req per sec"
	echo "[*] Flushing all data"
	bash remote_exec_sync.sh $redishost flushall.sh
	wait

	bash run_micro_ccmesh_launchers.sh $jobid
	wait
	
	bash remote_exec_sync.sh $schedhost run_wrk2.sh $rps
	wait

	scp root@${workers[0]}:~/ccmesh-server/monitor.txt data/${rps}_monitor.txt
done

scp root@$schedhost:~/data/* data/

echo "Done!"
