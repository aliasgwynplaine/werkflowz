#!/bin/bash
set -e

lhost=$(hostname)
cd ~/ccmesh-server
cargo build --release --bin hz-server 2> $HOME/ccmesh_build.log
echo "[$lhost]: Done!"

