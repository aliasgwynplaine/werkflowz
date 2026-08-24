#!/bin/bash
set -e
lhost=$(hostname)
cd ~/snitch
export GATEWAY_ADDR="$lhost:8080"
cargo run --release > $HOME/snitch.log
echo "[$lhost]: Done !"


