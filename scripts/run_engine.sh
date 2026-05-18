#!/bin/bash
# this script is meant to be executed from a remote host

set -e

if [ $# -lt 3 ]; then
	echo "usage: $0 <func_config.json> <node_id> <nightcore_gw_addr>" 
	exit 1
fi

source common.env # fix this

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

func_config=$1
node_id=$2
nightcore_gw_addr=$3

echo "rm -rf $BASE_DIR/outputs"
echo "mkdir -p $BASE_DIR/outputs"

echo "$NIGHTCORE_ROOT/bin/release/engine \
    --gateway_addr=$nightcore_gw_addr \
    --func_config_file=$BASE_DIR/$func_config \
    --node_id=$node_id \
    --v=1 2>$BASE_DIR/outputs/engine.log"
