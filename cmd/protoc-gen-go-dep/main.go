package main

import (
	"google.golang.org/protobuf/compiler/protogen"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/pluginpb"

    "fmt"
    "os"
    "io/ioutil"
    "strings"

)

type Generator struct {
    plugin *protogen.Plugin
    write bool
    messages map[string]struct{}
    suppressWarn bool
}

func NewGenerator(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*Generator, error) {
    plugin, err := opts.New(request)
    if err != nil {
        return nil, err
    }

    generator := &Generator{
        plugin: plugin,
        messages: make(map[string]struct{}),
        suppressWarn: false,
    }
    
    params := parseParameter(request.GetParameter())

    if _, ok := params["quiet"]; ok {
		generator.suppressWarn = true
	}

    return generator, nil
}

func (p *Generator) Name() string {
    return "dep"
}

func (p *Generator) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	genFileMap := make(map[string]*protogen.GeneratedFile)
    
    for _, protoFile := range p.plugin.Files {
        fileName := protoFile.GeneratedFilenamePrefix + ".pb.dep.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")
		genFileMap[fileName] = g


        g.P("package ", protoFile.GoPackageName)

        for _, message := range protoFile.Messages {
            p.generateDepsFunction(g, message)
        }

    }

    return p.plugin.Response(), nil
}

func (p *Generator) generateDepsFunction(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())
	msgName := string(message.Desc.Name())

    g.P(`// Deps function returns a static string for the time being, needs dev`)
	g.P(`func (t *`, typeName, `) Deps  () string {`)

    g.P(`return "`, msgName, `"`)
	g.P(`}`)
}

func main() {
    input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		panic(err)
	}

	opts := protogen.Options{}

    generator, err := NewGenerator(opts, &request)
    if err != nil {
        panic(err)
    }

    response, err := generator.Generate()
    if err != nil {
        panic(err)
    }

    out, err := proto.Marshal(response)
    if err != nil {
        panic(err)
    }

    fmt.Fprint(os.Stdout, string(out))

/*	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})*/
}

func parseParameter(param string) map[string]string {
	paramMap := make(map[string]string)

	params := strings.Split(param, ",")
	for _, param := range params {
		if strings.Contains(param, "=") {
			kv := strings.Split(param, "=")
			paramMap[kv[0]] = kv[1]
			continue
		}
		paramMap[param] = ""
	}

	return paramMap
}

// generateFile generates a _dep.pb.go file containing gRPC service definitions.
/*func generateFile(gen *protogen.Plugin, file *protogen.File) {
	filename := file.GeneratedFilenamePrefix + "_dep.pb.go"
    t := template.Must(template.New("dep").Parse(tmpl))
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
    var buf bytes.Buffer
    t.Execute(&buf, gen)

	g.P("// Code generated by protoc-gen-go-dep. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P("import ", "\"fmt\"")
    log.Println(buf.String())

//	return g
}*/
