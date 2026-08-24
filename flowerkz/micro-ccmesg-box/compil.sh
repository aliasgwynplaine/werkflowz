#!/bin/bash
set -e

echo "Building..."

go build -o Entry1 src/main.go
cp Entry1 Entry2
cp Entry1 Sink

echo "Done!"
