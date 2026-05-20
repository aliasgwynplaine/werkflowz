#!/bin/bash
set -e

echo "Building..."

go build incrementor.go
go build incrementorFinish.go
go build incrementorInit.go

echo "Done!"
