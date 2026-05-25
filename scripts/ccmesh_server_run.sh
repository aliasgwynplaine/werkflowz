#!/bin/bash
set -e

if [ $# -lt 1 ]; then
	echo "fuck you"
	exit 1
fi

lhost=$(hostname)
cd ~/ccmesh-server
cargo run --release --bin hz-server $1 2> $HOME/ccmesh_server.log
echo "[$lhost]: Done!"

