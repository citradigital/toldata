#!/bin/sh

ls -l /src
apk add --update git binutils
cd /src
go mod init protonats
CGO_ENABLED=0 GOOS=linux go build -o /build/protoc-gen-protonats 
strip /build/protoc-gen-protonats