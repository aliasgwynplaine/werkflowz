#!/bin/bash
set -e 

BASE_DIR=$(realpath $(dirname $0))
NIGHTCORE_ROOT=../../nightcore
BUILD_TYPE=release

rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

export NIGHTCORE_GW_ADDR=127.0.0.1

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
    --fprocess=$BASE_DIR/init \
    --v=1 2>$BASE_DIR/outputs/launcher_init.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=2 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/writer \
    --v=1 2>$BASE_DIR/outputs/launcher_writer.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=3 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/reader \
    --v=1 2>$BASE_DIR/outputs/launcher_reader.log &

wait
