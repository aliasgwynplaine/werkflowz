#!/bin/bash
set -e

echo "Building..."

go build src/fanin/fanin.go
go build src/fanout/fanout.go
go build src/incrementor/incrementor.go

echo "Done!"