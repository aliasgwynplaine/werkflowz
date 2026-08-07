#!/bin/bash
set -e

if [ $# -lt 1 ]; then
	echo "fuck you"
	exit 1
fi

lhost=$(hostname)
cd ~/snitch
cargo run --release > $HOME/snitch.log
echo "[$lhost]: Done !"


