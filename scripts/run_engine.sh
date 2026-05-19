#!/bin/bash
# this script is meant to be executed from a remote host

set -e

if [ $# -lt 3 ]; then
	echo "usage: $0 <experiment-name> <node_id> <gateway_addr>" 
	exit 1
fi

lhost=$(hostname)

if [ ! -d "$HOME/$1" ]; then
	echo "[$lhost]: Experiment not loaded"
	exit 42
fi

echo "[$lhost]: deploying engine for experiment $1"

experiment=$1
node_id=$2
nightcore_gw_addr=$3

cd $HOME/$experiment
source common.env

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

test -d $BASE_DIR/outputs && rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

$NIGHTCORE_ROOT/bin/release/engine \
    --gateway_addr=$nightcore_gw_addr \
    --func_config_file=$BASE_DIR/func_config.json \
    --node_id=$node_id \
    --v=1 2>$BASE_DIR/outputs/engine.log &

echo "[$lhost]: engine has been deployed for experiment $experiment."
