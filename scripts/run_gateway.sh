#!/bin/bash
# this script is meant to be executed from a remote host

set -e

if [ $# -lt 1 ]; then
	echo "usage: $0 <experiment-name> "
	exit 1
fi

lhost=$(hostname)

if [ ! -d "$HOME/flowerkz/$1" ]; then
	echo "[$lhost]: Experiment not loaded"
	exit 42
fi

experiment=$1
cd $HOME/flowerkz/$experiment
source common.env # fix this

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

[ -d "$BASE_DIR/outputs" ] && rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

ulimit -n 1000000
echo 1024 65535 > /proc/sys/net/ipv4/ip_local_port_range

$NIGHTCORE_ROOT/bin/release/gateway \
	--listen_addr=0.0.0.0 \
	--lb_pick_least_load \
	--num_io_workers=64 \
	--max_running_requests=128 \
	--func_config_file=$BASE_DIR/func_config.json \
	--v=1 2> $BASE_DIR/outputs/gateway.log &

echo "[$lhost]: gateway has been deployed for experiment $experiment"

