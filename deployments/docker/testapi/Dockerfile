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

FROM golang:1.19-alpine
ENV _DBHOST testdb
ENV _DBPORT 5432
ENV NATS_URL nats://testnats:4222
ENV CONFIG_FILE /config.json
ENV GO111MODULE on
ENV COMPONENT ${COMPONENT}
ENV WORKDIR /go/src/${COMPONENT}
RUN apk update
RUN apk add curl git gcc musl-dev bash jq
RUN mkdir -p ${WORKDIR}/${COMPONENT}
WORKDIR ${WORKDIR}
COPY *.go ${WORKDIR}/${COMPONENT}/
RUN go mod init toldata
COPY configs/config.json /config.json
COPY deployments/docker/testapi/start.sh /
RUN chmod +x /start.sh
WORKDIR ${WORKDIR}/${COMPONENT}
