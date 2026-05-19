#!/bin/bash
# this script is meant to be executed from a remote host

set -e

if [ $# -lt 1 ]; then
	echo "usage: $0 <experiment-name> "
	exit 1
fi

lhost=$(hostname)

if [ ! -d "$HOME/$1" ]; then
	echo "[$lhost]: Experiment not loaded"
	exit 42
fi

experiment=$1
cd $HOME/$experiment
source common.env # fix this

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

[ -d "$BASE_DIR/outputs" ] && rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

$NIGHTCORE_ROOT/bin/release/gateway \
    --func_config_file=$BASE_DIR/func_config.json \
    --v=1 2>$BASE_DIR/outputs/gateway.log &

echo "[$lhost]: gateway has been deployed for experiment $experiment"

