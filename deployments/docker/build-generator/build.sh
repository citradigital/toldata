#!/bin/sh

ls -l /src
apk update && apk add git
cd /src
go mod init protonats
go build -o /build/protoc-gen-protonats 