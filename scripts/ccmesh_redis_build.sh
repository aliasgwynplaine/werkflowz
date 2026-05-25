#!/bin/bash
set -e

lhost=$(hostname)
cd ~/ccmesh-server
echo "[$lhost]: Building redis ccmesh server..."
cargo build --release --bin redis 2> $HOME/redis_build.log
echo "[$lhost]: Done!"
echo "[$lhost]: Running redis ccmesh server..."
cargo run --release --bin redis 2> $HOME/redis.log

