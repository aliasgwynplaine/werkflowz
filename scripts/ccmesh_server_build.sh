#!/bin/bash
set -e

lhost=$(hostname)
cd ~/ccmesh-server
cargo build --release --bin hz-server
echo "[$lhost]: Done!"

