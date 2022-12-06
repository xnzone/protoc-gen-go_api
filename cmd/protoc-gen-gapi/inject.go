package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateInject(_ *protogen.Plugin, fd *protogen.File) {
	fileName := strings.ReplaceAll(fd.Proto.GetName(), ".proto", "pb.go")
	if strings.Contains(*fd.Proto.GetOptions().GoPackage, "/") {
		names := strings.Split(fd.Proto.GetName(), "/")
		fileName = path.Join(*fd.Proto.GetOptions().GoPackage, strings.ReplaceAll(names[len(names)-1], ".proto", ".pb.go"))
	}
	text, err := parseFile(fileName, nil)
	if err != nil {
		log.Println("generate inject err", err)
		return
	}
	if err = inject(fileName, text); err != nil {
		log.Println("generate inject err", err)
		return
	}
}

var (
	reginject = regexp.MustCompile("`.+`$")
	regtags   = regexp.MustCompile(`[\w_]+:"[^"]+"`)
)

type textArea struct {
	Start      int
	End        int
	CurrentTag string
	InjectTag  string
}

func parseFile(inputPath string, xxxSkip []string) (areas []textArea, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inputPath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	for _, decl := range f.Decls {
		// check if is generic declaration
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		var typeSpec *ast.TypeSpec
		for _, spec := range genDecl.Specs {
			if ts, tsOK := spec.(*ast.TypeSpec); tsOK {
				typeSpec = ts
				break
			}
		}

		// skip if can't get type spec
		if typeSpec == nil {
			continue
		}

		// not a struct, skip
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		builder := strings.Builder{}
		if len(xxxSkip) > 0 {
			for i, skip := range xxxSkip {
				builder.WriteString(fmt.Sprintf("%s:\"-\"", skip))
				if i > 0 {
					builder.WriteString(",")
				}
			}
		}

		for _, field := range structDecl.Fields.List {
			if field == nil || field.Tag == nil {
				continue
			}
			if len(field.Names) <= 0 {
				continue
			}

			currentTag := field.Tag.Value
			tagbeg := strings.Index(field.Tag.Value, "json:")
			if tagbeg == -1 {
				continue
			}
			tagend := strings.Index(field.Tag.Value, ",omitempty")
			if tagend == -1 {
				tagend = len(field.Tag.Value) - 2
			}
			tagName := field.Tag.Value[tagbeg+len(`json:"`) : tagend]
			tags := make([]string, 0, 4)
			if strings.Contains(field.Comment.Text(), "http.header") {
				tags = append(tags, fmt.Sprintf(`header:"%s"`, tagName))
			}
			if strings.Contains(field.Comment.Text(), "http.uri") {
				tags = append(tags, fmt.Sprintf(`uri:"%s"`, tagName))
			}
			if strings.Contains(field.Comment.Text(), "http.query") {
				tags = append(tags, fmt.Sprintf(`form:"%s"`, tagName))
			}
			tag := strings.Join(tags, " ")

			area := textArea{
				Start:      int(field.Pos()),
				End:        int(field.End()),
				CurrentTag: currentTag[1 : len(currentTag)-1],
				InjectTag:  tag,
			}
			areas = append(areas, area)
		}
	}
	return
}

func inject(inputPath string, areas []textArea) (err error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	if err = f.Close(); err != nil {
		return
	}

	// inject custom tags from tail of file first to preserve order
	for i := range areas {
		area := areas[len(areas)-i-1]
		contents = injectTag(contents, area)
	}
	if err = ioutil.WriteFile(inputPath, contents, 0644); err != nil {

		return
	}

	return
}

type tagItem struct {
	key   string
	value string
}

type tagItems []tagItem

func (ti tagItems) format() string {
	tags := make([]string, 0)
	for _, item := range ti {
		tags = append(tags, fmt.Sprintf(`%s:%s`, item.key, item.value))
	}
	return strings.Join(tags, " ")
}

func (ti tagItems) override(nti tagItems) tagItems {
	overrided := make([]tagItem, 0)
	for i := range ti {
		var dup = -1
		for j := range nti {
			if ti[i].key == nti[j].key {
				dup = j
				break
			}
		}
		if dup == -1 {
			overrided = append(overrided, ti[i])
		} else {
			overrided = append(overrided, nti[dup])
			nti = append(nti[:dup], nti[dup+1:]...)
		}
	}
	return append(overrided, nti...)
}

func newTagItems(tag string) tagItems {
	items := make([]tagItem, 0)
	splitted := regtags.FindAllString(tag, -1)

	for _, t := range splitted {
		sepPos := strings.Index(t, ":")
		items = append(items, tagItem{
			key:   t[:sepPos],
			value: t[sepPos+1:],
		})
	}
	return items
}

func injectTag(contents []byte, area textArea) (injected []byte) {
	expr := make([]byte, area.End-area.Start)
	copy(expr, contents[area.Start-1:area.End-1])
	cti := newTagItems(area.CurrentTag)
	iti := newTagItems(area.InjectTag)
	ti := cti.override(iti)
	expr = reginject.ReplaceAll(expr, []byte(fmt.Sprintf("`%s`", ti.format())))
	injected = append(injected, contents[:area.Start-1]...)
	injected = append(injected, expr...)
	injected = append(injected, contents[area.End-1:]...)
	return
}
