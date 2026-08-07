#!/bin/bash
set -e

echo "Building..."

go build src/writer/writer.go
go build src/readerwriter/readerwriter.go
go build src/finalreader/finalreader.go

echo "Done!"