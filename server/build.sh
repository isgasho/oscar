#! /bin/bash

set -e

export BUILD_TIME=`date -u +%Y-%m-%d-%I:%M:%S`
# https://github.com/golang/go/issues/26492
go build -tags 'osusergo netgo' -ldflags '-X main.ServerBuildTime=$BUILD_TIME -extldflags "-fno-PIC -static" -s -w' -buildmode pie -o oscar
