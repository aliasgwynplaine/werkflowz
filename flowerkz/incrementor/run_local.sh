#!/bin/bash

BASE_DIR=$(realpath $(dirname $0))
NIGHTCORE_ROOT=../nightcore
BUILD_TYPE=release

rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

export NIGHTCORE_GW_ADDR=127.0.0.1 # this must change for the distributed case

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/gateway \
    --func_config_file=$BASE_DIR/func_config.json \
    --v=1 2>$BASE_DIR/outputs/gateway.log &

sleep 1

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/engine \
    --func_config_file=$BASE_DIR/func_config.json \
    --node_id=0 \
    --v=1 2>$BASE_DIR/outputs/engine.log &

sleep 1

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=1 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/incrementorInit \
    --v=1 2>$BASE_DIR/outputs/launcher_incrementorInit.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=2 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/incrementor \
    --v=1 2>$BASE_DIR/outputs/launcher_incrementor.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=3 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/incrementorFinish \
    --v=1 2>$BASE_DIR/outputs/launcher_incrementorFinish.log &

wait
