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

include .env
export $(shell sed 's/=.*//' .env)

IMAGE_TAG ?= master
export $IMAGE_TAG

.PHONY : test

test: 
	MIGRATION_PATH=${MIGRATION_PATH} cd test && go test

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
	docker run -v `pwd`:/gen -v `pwd`/api:/api citradigital/protonats -I /api/ /api/nats.proto --gogofast_out=/gen
	docker run -v `pwd`/test:/gen -v `pwd`/api:/api citradigital/protonats -I /api/ /api/nats_test.proto --protonats_out=/gen --gogofast_out=/gen

generator:
	go build -o protonats-gen cmd/protonats-gen/main.go

build-generator:
	mkdir -p tmp
	cp -a cmd/protonats-gen tmp
	docker run -v `pwd`/deployments/docker/build:/build -v `pwd`/tmp/protonats-gen:/src -v `pwd`/deployments/docker/build-generator/build.sh:/build.sh golang:1.12-alpine /build.sh
	docker build -t citradigital/protonats -f deployments/docker/build-generator/Dockerfile deployments/docker/