#!/bin/bash
# this script is meant to be executed from a remote host
# and also meant to be executed in a container... ?

set -e

if [ $# -lt 4 ]; then
	echo "usage: $0 <func_id> <fprocess_mode> <fprocess>" 
	exit 1
fi

source common.env # fix this

if [ -z "$COMMON_ENV_H" ]; then
    echo '$COMMON_ENV_H is not defined'
    exit 2
fi

func_id=$1
fprocess_mode=$2
fprocess=$3

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=$func_id --fprocess_mode=$fprocess_mode \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/$fprocess \
    --v=1 2>$BASE_DIR/outputs/launcher_$fprocess.log &