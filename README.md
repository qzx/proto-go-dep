# Basic structure for a protoc plugin to generate code from proto as a protoc plugin

```shell
$ cd cmd/protoc-gen-go-dep/
$ go install .
```

Once This is installed you can use:

```shell
$ protoc --go_out=. --go_opt=paths=source_relative --go-dep_out=. --go-dep_opt=paths=source_relative example/example.proto
```

to generate protobuf structs as well as our deps file. Currently it only creates a single function for each message type that returns a static string:

See the function:

```go
func (p *Generator) generateDepsFunction(g *protogen.GeneratedFile, message *protogen.Message) {
        typeName := string(message.Desc.Name())
        msgName := string(message.Desc.Name())

        g.P(`// Deps function returns a static string for the time being, needs dev`)
        g.P(`func (t *`, typeName, `) Deps  () string {`)

    g.P(`return "`, msgName, `"`)
        g.P(`}`)
}
```



