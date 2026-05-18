#!/bin/bash
set -x

echo "Building..."

go build incrementor.go
go build incrementorFinish.go
go build incrementorInit.go

echo "Done!"
