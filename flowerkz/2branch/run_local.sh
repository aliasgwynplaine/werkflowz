#!/bin/bash
set -e 

BASE_DIR=$(realpath $(dirname $0))
NIGHTCORE_ROOT=../../nightcore
SNITCH_ROOT=../../snitch
BUILD_TYPE=release

rm -rf $BASE_DIR/outputs
mkdir -p $BASE_DIR/outputs

export NIGHTCORE_GW_ADDR=127.0.0.1
#export GATEWAY_ADDR="127.0.0.1:8080"
cd $SNITCH_ROOT
cargo run --release > $BASE_DIR/outputs/snitch.log &
cd -

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
    --fprocess=$BASE_DIR/fanout \
    --v=1 2>$BASE_DIR/outputs/launcher_fanout.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=2 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/incrementor \
    --v=1 2>$BASE_DIR/outputs/launcher_incrementor0.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=3 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/incrementor \
    --v=1 2>$BASE_DIR/outputs/launcher_incrementor1.log &

$NIGHTCORE_ROOT/bin/$BUILD_TYPE/launcher \
    --func_id=4 --fprocess_mode=go \
    --fprocess_output_dir=$BASE_DIR/outputs \
    --fprocess=$BASE_DIR/fanin \
    --v=1 2>$BASE_DIR/outputs/launcher_fanin.log &

echo "Ready!"
wait
    
