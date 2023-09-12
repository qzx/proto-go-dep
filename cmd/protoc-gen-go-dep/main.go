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
        g.P("import (")
        g.P(`   "database/sql"`)
        g.P(`   _ "github.com/lib/pq"`)
        g.P(")")
        g.P("")

        for _, message := range protoFile.Messages {
            p.generateModel(g, message)
            p.generateListFunction(g, message)
            p.generateCreateFunction(g, message)
            p.generateUpdateFunction(g, message)
            p.generateDeleteFunction(g, message)
            p.generateFormHandler(g, message)
            p.generateDepsFunction(g, message)
        }

    }

    return p.plugin.Response(), nil
}

func (p *Generator) generateModel(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// Lets start by creating a Model and Handler for our flow")
    g.P("type Handler struct {")
    g.P("   m *model")
    g.P("}")
    g.P("")
    g.P("type model struct {")
    g.P("   DB *sql.DB")
    g.P("   Table string")
    g.P("}")
    g.P("")
    g.P("func NewHandler(db *sql.DB) *Handler {")
    g.P("   model := New(model)")
    g.P("   model.DB = db")
    g.P(`   model.Table = "`, strings.ToLower(typeName), `"`) 
    g.P("}")
    g.P("")
}

func (p *Generator) generateListFunction(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// List function should return a list of these objects")
    g.P("func (m *model) List(tenant string) (map[int]", typeName, ", error) {")
    g.P("   ret := make(map[int]", typeName, ")")
    g.P("")
    g.P(`   rows, err := m.DB.Query("SELECT id, data FROM list_data($1, $2)", tenant, m.Table)`)
    g.P("   if err != nil { return ret, err }")
    g.P("")
    g.P("   defer rows.Close()")
    g.P("")
    g.P("   for rows.Next() {")
    g.P("       var row ", typeName)
    g.P("       var id int")
    g.P("")
    g.P("       err := rows.Scan(&id, &row)")
    g.P("       if err != nil { return ret, err }")
    g.P("")
    g.P("       ret[id] = row")
    g.P("   }")
    g.P("")
    g.P("   return ret, nil")
    g.P("}")
    g.P("")
}

func (p *Generator) generateGetFunction(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// Get function acquires a single record based on ID in database")
    g.P(`func (m *model) Get(tenant string, id int) (*`, typeName, `, error) {`)
    g.P("   ret := new(", typeName, ")")
    g.P("")
    g.P(`   err := m.DB.QueryRow("SELECT data FROM list_data($1, $2) WHERE id = $3",`)
    g.P("       tenant, m.Table, id).Scan(&contact)")
    g.P("")
    g.P("   if err != nil { return nil, err }")
    g.P("")
    g.P("   return ret, nil")
    g.P("}")
    g.P("")
}

func (p *Generator) generateCreateFunction(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())
    var firstField string
    if len(message.Fields) > 0 {
        firstField = message.Fields[0].GoName
    } else {
        return
    }

    checkField := strings.Join([]string{typeName, firstField}, ".")
    
    g.P("// Create function will create a new object of this type")
    g.P(`func (m *model) Create(tenant string, data *`, typeName, `) error {`)
    g.P("   if ", checkField, ` != "" {`)
    g.P(`       _, err := m.DB.Exec("CALL insert_data($1, $2, $3)", tenant, m.Table, contact)`)
    g.P("")
    g.P("       if err != nil { return err }")
    g.P("   } else {")
    g.P(`       return errors.New("`, firstField, ` Was not set")`)
    g.P("   }")
    g.P("")
    g.P("   return nil")
    g.P("}")
    g.P("")
}


func (p *Generator) generateUpdateFunction(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// Update function will replace the object stored at the given ID")
    g.P(`func (m *model) Update(tenant string, id string, data *`, typeName, `) error {`)
    g.P(`   _, err := m.DB.Exec("CALL update_data($1, $2, $3, $4)",`)
    g.P("       tenant, m.Table, id, data)")
    g.P("")
    g.P("   return err")
    g.P("}")
    g.P("")
}

func (p *Generator) generateDeleteFunction(g *protogen.GeneratedFile, message *protogen.Message) {
    g.P("// Delete function will... well delete the object at given ID")
    g.P(`func (m *model) Delete(tenant string, id string) error {`)
    g.P(`   _, err := m.DB.Exec("CALL delete_data_by_id($1, $2, $3)",`)
    g.P("       tenant, m.Table, id)")
    g.P("")
    g.P("   return err")
    g.P("}")
    g.P("")
}

func (p *Generator) generateFormHandler(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// A simple function to handle a htmx form and populate the struct")
    g.P(`func (x *`, typeName, `) HandleForm(req *http.Request) error {`)
    if len(message.Fields) > 0 {
        for _, field := range message.Fields {
            formField := strings.Join([]string{typeName, field.GoName}, "__")
            g.P(`       x.`, field.GoName, ` = req.FormValue("`, formField, `")`)
        }
    }
    g.P("   return x.Validate()")
    g.P("}")
}

func (p *Generator) generateDepsFunction(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())

    g.P(`// Deps function returns a static string for the time being, needs dev`)
	g.P(`func (t *`, typeName, `) Deps() string {`)
    g.P(`return "`, typeName, `"`)
	g.P(`}`)
    g.P("")
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
