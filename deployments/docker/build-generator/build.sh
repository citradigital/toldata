#!/bin/sh

ls -l /src
apk update && apk add git
cd /src
mkdir -p /cache/apk
mkdir -p /cache/go
ln -s /cache/apk /etc/apk/cache
go mod init protonats
go build -o /build/protoc-gen-protonats 