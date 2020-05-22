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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"

	"github.com/gogo/protobuf/proto"
)

func getServiceOption(options *descriptor.ServiceOptions, index int) string {
	descs, err := proto.ExtensionDescs(options)
	if err == nil {
		for _, desc := range descs {
			if desc.Field == int32(index) {
				ext, err := proto.GetExtension(options, desc)
				if err == nil {
					bytes, ok := ext.([]byte)
					if ok {
						op, len := proto.DecodeVarint(bytes)
						tag := op >> 3
						wire := op & 7

						if wire == 2 && tag == uint64(index) {
							strlen := int(bytes[len])
							val := string(bytes[len+1 : strlen+len+1])
							return val
						}
						break
					}
				}
				break
			}
		}
	}
	return "/api"
}

func stripLastDot(name, packageName string) string {
	if packageName != "" && strings.HasPrefix(name, "."+packageName) {
		pos := strings.LastIndex(name, ".")
		if pos == -1 {
			return name
		}
		return name[pos+1:]
	}

	pos := strings.LastIndex(name, ".")
	if pos == -1 {
		return name
	}
	lpos := pos
	pos = strings.LastIndex(name[:lpos], ".")
	if pos == -1 {
		return name[lpos+pos+1:]
	}

	return name[pos+1:]
}

func stringPtr(in string) *string {
	if in == "" {
		return nil
	}

	return &in
}

func newTemplate(content string) (*template.Template, error) {
	fn := map[string]interface{}{
		"stripLastDot":     stripLastDot,
		"getServiceOption": getServiceOption,
	}

	return template.New("page").Funcs(fn).Parse(content)
}

func generateBase(in *descriptor.FileDescriptorProto, outputFormat, templateString string) (*plugin_go.CodeGeneratorResponse_File, error) {
	topPackageName := in.GetPackage()
	packageName := in.Options.GetGoPackage()
	if packageName == "" {
		return nil, errors.New("Unable to find go_package options in .proto file")
	}

	if topPackageName == "" {
		return nil, errors.New("Unable to find package declaration in .proto file")
	}

	buf := bytes.NewBuffer(nil)
	t, err := newTemplate(templateString)
	if err != nil {
		return nil, err
	}
	err = t.Execute(buf, map[string]interface{}{
		"File":        *in.Name,
		"PackageName": packageName,
		"Services":    in.Service,
		"Namespace":   topPackageName,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	filename := *in.Name
	ext := filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(ext)]
	filename = fmt.Sprintf(outputFormat, filename)

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    &filename,
		Content: stringPtr(buf.String()),
	}, nil

}

func generate(in *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {
	return generateBase(in, "%v.toldata.pb.go", rpcTemplate)
}

func generateGRPC(in *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {
	return generateBase(in, "%v.grpc.pb.go", grpcTemplate)
}

func generateREST(in *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {
	return generateBase(in, "%v.rest.pb.go", restTemplate)
}

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
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

		single, err := generate(file)
		if err != nil {
			log.Fatalln(err)
		}

		results = append(results, single)

		if strings.Contains(req.GetParameter(), "grpc") {
			single, err := generateGRPC(file)
			if err != nil {
				log.Fatalln(err)
			}

			results = append(results, single)
		}
		if strings.Contains(req.GetParameter(), "rest") {
			single, err := generateREST(file)
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
