#!/bin/bash
set -e

lhost=$(hostname)
cd ~/ccmesh-server
echo "[$lhost]: Building redis ccmesh server..."
cargo build --release --bin redis
echo "[$lhost]: Done!"
echo "[$lhost]: Running redis ccmesh server..."
cargo run --release --bin redis 2>&1 > $HOME/redis.log
echo "[$lhost]: Ready!"

