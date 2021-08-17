#!/bin/bash

cd $(dirname $BASH_SOURCE)

echo "Testing cryptolib"

#go clean -testcache
go test ./...
