package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func generateGAPI(_ *protogen.Plugin, fd *protogen.File) {
	data := getProtoFile(fd)
	bs, err := MakeGAPIContent(data)
	if err != nil {
		log.Println("make gapi content err: ", err)
		return
	}
	fileName := strings.ReplaceAll(fd.Proto.GetName(), ".proto", "pb.gapi.go")
	if strings.Contains(*fd.Proto.GetOptions().GoPackage, "/") {
		names := strings.Split(fd.Proto.GetName(), "/")
		fileName = path.Join(*fd.Proto.GetOptions().GoPackage, strings.ReplaceAll(names[len(names)-1], ".proto", ".pb.gapi.go"))
	}
	gfd, err := os.Create(fileName)
	if err != nil {
		log.Println("create gapi file err:", err)
		return
	}
	_, _ = fmt.Fprint(gfd, string(bs))
}

func getProtoFile(fd *protogen.File) *ProtoFile {
	pf := new(ProtoFile)
	pf.Source = fd.Proto.GetName()
	pf.Package = string(fd.GoPackageName)
	pf.Version = Version
	for _, srv := range fd.Services {
		pf.Services = append(pf.Services, getProtoService(srv))
	}
	return pf
}

func getProtoService(srv *protogen.Service) *ProtoService {
	ps := new(ProtoService)
	ps.Name = strings.ReplaceAll(srv.GoName, "Service", "")
	ps.Comment = srv.Comments.Leading.String()
	for _, m := range srv.Methods {
		ps.Methods = append(ps.Methods, getProtoMethod(m))
	}
	return ps
}

func getProtoMethod(m *protogen.Method) *ProtoMethod {
	pm := new(ProtoMethod)
	pm.Name = m.GoName
	pm.Comment = m.Comments.Leading.String()
	pm.ServiceName = m.Parent.GoName
	pm.RequestType = m.Input.GoIdent.GoName
	pm.ResponseType = m.Output.GoIdent.GoName
	var body string
	pm.HTTPMethod, pm.RelativePath, body = buildMethod(extraMethodOptions(m))
	switch body {
	case "":
		pm.HasBody = false
	case "*":
		pm.HasBody = true
	default:
		pm.HasBody = true
		pm.Body = fmt.Sprintf(".%s", strcase.ToCamel(body))
	}
	return pm
}

// extraMethodOptions 解析proto的method的option
func extraMethodOptions(m *protogen.Method) *annotations.HttpRule {
	serverName := m.Parent.GoName
	rule := &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{Post: filepath.Join("/", serverName, m.GoName)},
	}
	if m.Desc.Options() == nil {
		return rule
	}

	ext := proto.GetExtension(m.Desc.Options(), annotations.E_Http)
	opts, ok := ext.(*annotations.HttpRule)
	if !ok {
		log.Print("assert extension")
		return rule
	}
	return opts
}

// buildMethod 把httpRule转成http的method和uri
func buildMethod(rule *annotations.HttpRule) (method string, path string, body string) {
	body = rule.Body
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "GET"
		return
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
		return
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "POST"
		if body == "" {
			body = "*"
		}
		return
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
		return
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
		return
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
		return
	default:
		path = pattern.(*annotations.HttpRule_Post).Post
		method = "POST"
		if body == "" {
			body = "*"
		}
		return
	}
}
