package toldata

import (
	"bytes"
	"errors"
	fmt "fmt"
	"go/format"
	"html/template"
	"log"
	"path/filepath"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

type Generator struct {
	File *descriptor.FileDescriptorProto
}

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
		return nil, errors.New("unable to find go_package options in .proto file")
	}

	if topPackageName == "" {
		return nil, errors.New("unable to find package declaration in .proto file")
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

	s, err := format.Source(buf.Bytes())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    &filename,
		Content: stringPtr(string(s)),
	}, nil

}

func (g Generator) Generate() (*plugin_go.CodeGeneratorResponse_File, error) {
	return generateBase(g.File, "%v.toldata.pb.go", RPCTemplate)
}

func (g Generator) GenerateGRPC() (*plugin_go.CodeGeneratorResponse_File, error) {
	return generateBase(g.File, "%v.grpc.pb.go", GRPCTemplate)
}

func (g Generator) GenerateREST() (*plugin_go.CodeGeneratorResponse_File, error) {
	return generateBase(g.File, "%v.rest.pb.go", RestTemplate)
}
