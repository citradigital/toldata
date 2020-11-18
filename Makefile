# Copyright 2019 Citra Digital Lintas
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIR=deployments/docker
RECIPE=${DIR}/docker-compose.yaml
NAMESPACE=builder

DIND_PREFIX ?= $(HOME)
ifneq ($(HOST_PATH),)
DIND_PREFIX := $(HOST_PATH)
endif
ifeq ($(CACHE_PREFIX),)
	CACHE_PREFIX=/tmp
endif

PREFIX=$(shell echo $(PWD) | sed -e s:$(HOME):$(DIND_PREFIX):)

include .env
export $(shell sed 's/=.*//' .env)

IMAGE_TAG ?= latest
export $IMAGE_TAG

.PHONY : test

test: 
	docker run \
		--network ${NAMESPACE}_default \
		--env-file .env \
		-v $(CACHE_PREFIX)/cache/go:/go/pkg/mod \
		-v $(CACHE_PREFIX)/cache/apk:/etc/apk/cache \
		-v $(PREFIX)/deployments/docker/build:/build \
		-v $(PREFIX)/:/src \
		-v $(PREFIX)/scripts/test.sh:/test.sh \
		-e UID=$(UID) \
		golang:1.12-alpine /test.sh 

buildtest: 
	docker-compose -f ${RECIPE} -p ${NAMESPACE} build testapi

infratest:
	docker-compose -f ${RECIPE} -p ${NAMESPACE} up -d testnats

cleantest:
	docker-compose -f ${RECIPE} -p ${NAMESPACE} stop 
	docker-compose -f ${RECIPE} -p ${NAMESPACE} rm -f testapi
	docker-compose -f ${RECIPE} -p ${NAMESPACE} rm -f testnats


runtest: cleantest buildtest
	docker-compose -f ${RECIPE} -p ${NAMESPACE} run testapi /start.sh
	docker-compose -f ${RECIPE} -p ${NAMESPACE} down 

gen_clean:
	rm -f *.pb.go

gen: 
	docker run -v $(PREFIX):/gen -v $(PREFIX)/api:/api citradigital/toldata -I /api/ /api/toldata.proto --gogofaster_out=/gen
	docker run -v $(PREFIX)/test:/gen -v $(PREFIX)/api:/api citradigital/toldata -I /api/ /api/toldata_test.proto --toldata_out=plugins=rest,grpc:/gen --gogofaster_out=plugins=grpc:/gen

generator:
	go build -o toldata-gen cmd/toldata-gen/main.go cmd/toldata-gen/templates.go

build-generator:
	mkdir -p tmp/src
	cp -a *.go go.mod cmd tmp/src
	cp api/toldata.proto deployments/docker/build/
	docker run -v $(CACHE_PREFIX)/cache/go:/go/pkg/mod \
		-v $(CACHE_PREFIX)/cache/apk:/etc/apk/cache \
		-v $(PREFIX)/deployments/docker/build:/build \
		-v $(PREFIX)/tmp/src:/src \
		-v $(PREFIX)/deployments/docker/build-generator/build.sh:/build.sh \
		golang:1.12-alpine /build.sh
	docker build -t citradigital/toldata:$(IMAGE_TAG) -f deployments/docker/build-generator/Dockerfile deployments/docker/
