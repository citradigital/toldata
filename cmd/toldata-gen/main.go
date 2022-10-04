// Copyright 2019 Citra Digital Lintas
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/darmawan01/toldata"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"

	"github.com/gogo/protobuf/proto"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err)
	}

	req := plugin_go.CodeGeneratorRequest{}
	err = proto.Unmarshal(input, &req)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]*plugin_go.CodeGeneratorResponse_File, 0, len(req.ProtoFile))

	for _, file := range req.ProtoFile {
		if len(file.Service) == 0 {
			continue
		}

		gen := toldata.Generator{
			File: file,
		}

		single, err := gen.Generate()
		if err != nil {
			log.Fatalln(err)
		}

		results = append(results, single)

		if strings.Contains(req.GetParameter(), "grpc") {
			single, err := gen.GenerateGRPC()
			if err != nil {
				log.Fatalln(err)
			}

			results = append(results, single)
		}
		if strings.Contains(req.GetParameter(), "rest") {
			single, err := gen.GenerateREST()
			if err != nil {
				log.Fatalln(err)
			}

			results = append(results, single)
		}

	}

	res := &plugin_go.CodeGeneratorResponse{
		File: results,
	}
	result, err := proto.Marshal(res)
	if err != nil {
		log.Fatalln(err)
	}

	os.Stdout.Write(result)
}
