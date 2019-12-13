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

FROM herpiko/protoc

COPY build/protoc-gen-toldata /usr/bin
RUN mkdir -p /protobuf/github.com/citradigital/toldata
COPY build/toldata.proto /protobuf/github.com/citradigital/toldata

ENTRYPOINT ["/usr/bin/protoc", "-I/protobuf"]
