package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-gapi %v\n", Version)
	}
	protogen.Options{ParamFunc: flag.CommandLine.Set}.Run(gen)
}

func gen(p *protogen.Plugin) error {
	for _, fd := range p.Files {
		if !fd.Generate {
			continue
		}
		generateGAPI(p, fd)
		generateInject(p, fd)
	}
	return nil
}
