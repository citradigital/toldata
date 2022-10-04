#!/bin/sh

apk add --update git binutils

cd /src

go mod tidy
CGO_ENABLED=0 GOOS=linux go build -o /build/protoc-gen-toldata /src/cmd/toldata-gen/*.go 

strip /build/protoc-gen-toldata
