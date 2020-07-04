#!/bin/sh

apk add --update curl git gcc musl-dev
su -l nobody
cd /src/test
env
echo "Starting test...."
go test -run "REST/$1" -timeout 30m
go test -run "GRPC/$1" -timeout 30m
