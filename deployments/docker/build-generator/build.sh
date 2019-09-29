#!/bin/sh

ls -l /src
apk add --update git binutils
cd /src
go mod init toldata
CGO_ENABLED=0 GOOS=linux go build -o /build/protoc-gen-toldata 
strip /build/protoc-gen-toldata