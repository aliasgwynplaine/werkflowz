#!/bin/bash
# this script is meant to be executed from a remote host
# and also meant to be executed in a container... ?

set -e

if [ $# -lt 2 ]; then
	echo "usage: $0 <experiment-name> <func_id> <gateway_addr>" 
	exit 1
fi

lhost=$(hostname)

if [ ! -d "$HOME/$1" ]; then
	echo "[$lhost]: Experiment not loaded"
	exit 42
fi

experiment=$1
cd $HOME/$experiment
source common.env

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

func_id=$2

# all this info will be retrieved from the nighcore-config...
fprocess_mode=$(cat nightcore_config.json | jq -r ".fprocess_mode")
fprocess=$(cat nightcore_config.json | jq -r ".fuxmap[$((func_id - 1))]")

if [ "$fprocess" = "null" ]; then
	echo "[$lhost]: error bad func_id"
	exit 43
fi

export NIGHTCORE_GW_ADDR=$3

$NIGHTCORE_ROOT/bin/release/launcher \
    --func_id=$func_id --fprocess_mode=$fprocess_mode \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/$fprocess \
    --v=1 2> $BASE_DIR/outputs/launcher_$fprocess.log &

echo "[$lhost]: $fprocess_mode function $func_id - $fprocess  has been launched."
