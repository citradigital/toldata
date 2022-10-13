package toldata

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"

	descriptor "google.golang.org/protobuf/types/descriptorpb"
	plugin "google.golang.org/protobuf/types/pluginpb"
)

type Generator struct {
	File                *descriptor.FileDescriptorProto
	PathsSourceRelative bool
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

// baseName returns the last path element of the name, with the last dotted suffix removed.
func baseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[0:i]
	}
	return name
}

// getGoPackage returns the file's go_package option.
// If it containts a semicolon, only the part before it is returned.
func getGoPackage(fd *descriptor.FileDescriptorProto) string {
	pkg := fd.GetOptions().GetGoPackage()
	if strings.Contains(pkg, ";") {
		parts := strings.Split(pkg, ";")
		if len(parts) > 2 {
			log.Fatalf("protoc-gen-toldata: go_package '%s' contains more than 1 ';'", pkg)
		}
		pkg = parts[1]
	}

	return pkg
}

// goPackageOption interprets the file's go_package option.
// If there is no go_package, it returns ("", "", false).
// If there's a simple name, it returns ("", pkg, true).
// If the option implies an import path, it returns (impPath, pkg, true).
func goPackageOption(d *descriptor.FileDescriptorProto) (impPath, pkg string, ok bool) {
	pkg = getGoPackage(d)
	if pkg == "" {
		return
	}
	ok = true
	// The presence of a slash implies there's an import path.
	slash := strings.LastIndex(pkg, "/")
	if slash < 0 {
		return
	}
	impPath, pkg = pkg, pkg[slash+1:]
	// A semicolon-delimited suffix overrides the package name.
	sc := strings.IndexByte(impPath, ';')
	if sc < 0 {
		return
	}
	impPath, pkg = impPath[:sc], impPath[sc+1:]
	return
}

// goPackageName returns the Go package name to use in the
// generated Go file.  The result explicit reports whether the name
// came from an option go_package statement.  If explicit is false,
// the name was derived from the protocol buffer's package statement
// or the input file name.
func goPackageName(d *descriptor.FileDescriptorProto) (name string, explicit bool) {
	// Does the file have a "go_package" option?
	if _, pkg, ok := goPackageOption(d); ok {
		return pkg, true
	}

	// Does the file have a package clause?
	if pkg := d.GetPackage(); pkg != "" {
		return pkg, false
	}
	// Use the file base name.
	return baseName(d.GetName()), false
}

func (g Generator) generateBase(outputFormat, templateString string) (*plugin.CodeGeneratorResponse_File, error) {
	topPackageName := g.File.GetPackage()
	packageName := g.File.Options.GetGoPackage()
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

	cleanedPkgName, _ := goPackageName(g.File)
	err = t.Execute(buf, map[string]interface{}{
		"File":        *g.File.Name,
		"PackageName": cleanedPkgName,
		"Services":    g.File.Service,
		"Namespace":   topPackageName,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	code := strings.ReplaceAll(buf.String(), "&lt;", "<")
	s, err := format.Source([]byte(code))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	filename := *g.File.Name
	ext := path.Ext(filename)
	filename = filename[0 : len(filename)-len(ext)]
	filename = fmt.Sprintf(outputFormat, filename)

	return &plugin.CodeGeneratorResponse_File{
		Name:    &filename,
		Content: stringPtr(string(s)),
	}, nil
}

func (g Generator) Generate() (*plugin.CodeGeneratorResponse_File, error) {
	return g.generateBase("%v.toldata.pb.go", RPCTemplate)
}

func (g Generator) GenerateGRPC() (*plugin.CodeGeneratorResponse_File, error) {
	return g.generateBase("%v.grpc.pb.go", GRPCTemplate)
}

func (g Generator) GenerateREST() (*plugin.CodeGeneratorResponse_File, error) {
	return g.generateBase("%v.rest.pb.go", RestTemplate)
}
