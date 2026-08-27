#!/bin/bash
set -e

lhost=$(hostname)
cd ~/ccmesh-server
echo "[$lhost]: building ccmesh scheduler"
cargo run --release --bin scheduler 2> $HOME/ccmesh_scheduler.log
echo "[$lhost]: Done!"

