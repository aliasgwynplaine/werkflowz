#!/bin/bash
set -e 
go build init.go
go build writer.go
go build reader.go
echo "Done!"
