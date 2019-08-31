#!/bin/sh

apk add --update curl git gcc musl-dev
su -l nobody
cd /src/test
env
echo "Starting test...."
go test -test.parallel 4
