#!/bin/bash
# this script is meant to be executed from a remote host

set -e

if [ $# -lt 1 ]; then
	echo "usage: $0 <func_config.json> " 
	exit 1
fi

source common.env # fix this

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

func_config=$1

rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

$NIGHTCORE_ROOT/bin/release/gateway \
    --func_config_file=$BASE_DIR/$1 \
    --v=1 2>$BASE_DIR/outputs/gateway.log

