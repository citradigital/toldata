#!/bin/sh

ls -lR /src/cmd/toldata-gen
apk add --update git binutils
cd /src
CGO_ENABLED=0 GOOS=linux go build -o /build/protoc-gen-toldata /src/cmd/toldata-gen/*.go 
strip /build/protoc-gen-toldata
