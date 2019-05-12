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

const (
	rpcTemplate = `// Code generated by github.com/citradigital/protonats. DO NOT EDIT.
// source: {{ .File }}
package {{ .PackageName }}
import (
	"context"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/citradigital/protonats"
	nats "github.com/nats-io/go-nats"
)

{{ range .Services }}{{ $ServiceName := .Name }}
type {{ .Name }}Interface interface {
	{{ range .Method }}{{ .Name }}(ctx context.Context, req *{{ .InputType | stripPackage }}) (*{{ .OutputType | stripPackage }}, error)
{{ end }}
}
{{ end }}

{{ range .Services }}{{ $ServiceName := .Name }}
type {{ $ServiceName }}Client struct {
	Bus *protonats.Bus
}

type {{ $ServiceName }}Server struct {
	Bus *protonats.Bus
	Service {{ $ServiceName }}Interface
}

func New{{ $ServiceName }}Client(bus *protonats.Bus) * {{$ServiceName}}Client {
	s := &{{ $ServiceName }}Client{ Bus: bus }
	return s
}

func New{{ $ServiceName }}Server(bus *protonats.Bus, service {{ $ServiceName }}Interface) * {{$ServiceName}}Server {
	s := &{{ $ServiceName }}Server{ Bus: bus, Service: service }
	return s
}


{{ range .Method }}	
func (service *{{ $ServiceName }}Client) {{ .Name }}(ctx context.Context, req *{{ .InputType | stripPackage }}) (*{{ .OutputType | stripPackage }}, error) {
	functionName := "{{ $ServiceName }}/{{ .Name }}"
	
	reqRaw, err := proto.Marshal(req)

	result, err := service.Bus.Connection.RequestWithContext(ctx, functionName, reqRaw)
	if err != nil {
		return nil, err
	}

	if result.Data[0] == 0 {
		// 0 means no error
		p := &{{ .OutputType | stripPackage }}{}
		err = proto.Unmarshal(result.Data[1:], p)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		var pErr protonats.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return nil, errors.New(pErr.ErrorMessage)
		} else {
			return nil, err
		}
	}
}


{{ end }}
{{ end }}

{{ range .Services }}{{ $ServiceName := .Name }}

func (service *{{ $ServiceName }}Server) Subscribe{{ .Name }}() error {
	bus := service.Bus
	
	var error errors
	{{ range .Method }}	

	_, err = bus.Connection.QueueSubscribe("{{ $ServiceName }}/{{ .Name }}", "{{ $ServiceName }}", func(m *nats.Msg) {
		var input {{ .InputType | stripPackage }}
		err := proto.Unmarshal(m.Data, &input)
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}
		result, err := service.Service.{{ .Name }}(bus.Context, &input)

		if m.Reply != ""  {
			if err != nil {
				bus.HandleError(m.Reply, err)
			} else {
				raw, err := proto.Marshal(result)
				if err != nil {
					bus.HandleError(m.Reply, err)
				} else {
					zero := []byte{0}
					bus.Connection.Publish(m.Reply, append(zero, raw...))
				}
			}
		}

	})

	{{ end }}

	return err
}


{{ end }}`
)

func stripPackage(name string) string {
	pos := strings.LastIndex(name, ".")
	if pos == -1 {
		return name
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
		"stripPackage": stripPackage,
	}

	return template.New("page").Funcs(fn).Parse(content)
}

func generate(in *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {

	packageName := in.Options.GetGoPackage()
	if packageName == "" {
		return nil, errors.New("Unable to find go_package options in .proto file")
	}

	buf := bytes.NewBuffer(nil)
	t, err := newTemplate(rpcTemplate)
	if err != nil {
		return nil, err
	}

	t.Execute(buf, map[string]interface{}{
		"File":        *in.Name,
		"PackageName": packageName,
		"Services":    in.Service,
	})

	filename := *in.Name
	ext := filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(ext)]
	filename = fmt.Sprintf("%v.protonats.pb.go", filename)

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    &filename,
		Content: stringPtr(buf.String()),
	}, nil
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
