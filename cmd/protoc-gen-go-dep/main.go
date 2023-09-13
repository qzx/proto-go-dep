package main

import (
	"google.golang.org/protobuf/compiler/protogen"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/pluginpb"
    "google.golang.org/protobuf/types/descriptorpb"

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
        if fileHasOurOptions(protoFile) != true {
            continue
        }

        fileName := protoFile.GeneratedFilenamePrefix + ".pb.dep.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")
		genFileMap[fileName] = g


        g.P("package ", protoFile.GoPackageName)
        g.P("import (")
        g.P(`   "database/sql"`)
        g.P(`   _ "github.com/lib/pq"`)
        g.P(`   "net/http"`)
        g.P(`   "github.com/go-chi/chi/v5"`)
        g.P(")")
        g.P("")

        for _, message := range protoFile.Messages {
            if messageHasOurOptions(message) == false {
                continue
            }
            p.generateListFunction(g, message)
            p.generateGetFunction(g, message)
            p.generateCreateFunction(g, message)
            p.generateUpdateFunction(g, message)
            p.generateDeleteFunction(g, message)
            p.generateFormHandler(g, message)
            p.generateTableFunction(g, message)
        }

    }

    return p.plugin.Response(), nil
}

func fileHasOurOptions(file *protogen.File) bool {
    for _, message := range file.Messages {
        if messageHasOurOptions(message) == true {
            return true
        }
    }
    return false
}

func messageHasOurOptions(message *protogen.Message) bool {
    opts := message.Desc.Options().(*descriptorpb.MessageOptions)
    if proto.HasExtension(opts, dep.E_Opts) {
        val := proto.GetExtension(opts, dep.E_Opts).(string)
		if val == "htmx" {
			return true
		}
	}
	return false
}

func (p *Generator) generateModel(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// Lets start by creating a Model and Handler for our flow")
    g.P("type Handler struct {")
    g.P("   DB *sql.DB")
    g.P("   Parent string")
    g.P("   ID string")
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

    g.P("// ListHandler is our http handler that acquires and renders a list of objects")
    g.P(`func (x *`, typeName, `) ListHandler(w http.ResponseWriter, req *http.Request) {`)
    g.P(`   db, err := r.Context().Value("db").(*sql.DB)`)
    g.P(`   if err != nil { return }`)
    g.P("")
    g.P(`   tenant := chi.URLParam(req, "id")`)
    g.P(`   ret, err := x.List(db, tenant)`)
    g.P(`   if err != nil { return }`)
    g.P("")
    g.P(`   jsonData, err := json.Marshal(data)`)
    g.P(`   if err != nil { return }`)
    g.P("")
    g.P(`   w.Header().Set("Content-Type", "application/json")`)
    g.P(`   w.Write(jsonData)`)
    g.P("}")
    g.P("")
    g.P("")
    g.P("// List function should return a list of these objects")
    g.P(`func (x *`, typeName, `) List(db *sql.DB, tenant string) (map[int]`, typeName, `, error) {`)
    g.P("   ret := make(map[int]", typeName, ")")
    g.P("")
    g.P(`   rows, err := db.Query("SELECT id, data FROM list_data($1, $2)", tenant, x.TableName())`)
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
    g.P(`func (x *`, typeName, `) Get(db *sql.DB, tenant string, id string) error {`)
    g.P("")
    g.P(`   return db.QueryRow("SELECT data FROM list_data($1, $2) WHERE id = $3",`)
    g.P("       tenant, x.TableName(), id).Scan(x)")
    g.P("")
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
    g.P(`func (x *`, typeName, `) Create(db *sql.DB, tenant string, data *`, typeName, `) error {`)
    g.P("   if ", checkField, ` != "" {`)
    g.P(`       _, err := db.Exec("CALL insert_data($1, $2, $3)", tenant, x.TableName(), contact)`)
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
    g.P(`func (x *`, typeName, `) Update(db *sql.DB, tenant string, id string, data *`, typeName, `) error {`)
    g.P(`   _, err := db.Exec("CALL update_data($1, $2, $3, $4)",`)
    g.P("       tenant, x.TableName(), id, data)")
    g.P("")
    g.P("   return err")
    g.P("}")
    g.P("")
}

func (p *Generator) generateDeleteFunction(g *protogen.GeneratedFile, message *protogen.Message) {
    typeName := string(message.Desc.Name())

    g.P("// Delete function will... well delete the object at given ID")
    g.P(`func (x *`, typeName, `) Delete(db *sql.DB, tenant string, id string) error {`)
    g.P(`   _, err := db.Exec("CALL delete_data_by_id($1, $2, $3)",`)
    g.P("       tenant, x.TableName(), id)")
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

func (p *Generator) generateTableFunction(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())

    g.P(`// Deps function returns a static string for the time being, needs dev`)
	g.P(`func (*`, typeName, `) TableName() string {`)
    g.P(`   return "`, strings.ToLower(typeName), `"`)
	g.P(`}`)
    g.P("")
}

func (p *Generator) generateRouteFunction(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())

    g.P(`// Route function will return chi.Router that can be mounted to a parent router`)
    g.P(`func (x *`, typeName, `) Routes() chi.Router {`)
    g.P("   r := chi.NewRouter()")
    g.P("")
    g.P(`   r.Get("/", x.ListHandler)`)
    g.P(`   r.Post("/", x.CreateHandler)`)
    g.P(`   r.Route("/{contact}", func(r chi.Router) {`)
    g.P(`       r.Get("/", x.GetHandler)`)
    g.P(`       r.Put("/", x.UpdateHandler)`)
    g.P(`       r.Delete("/", x.DeleteHandler)`)
    g.P("   })")
    g.P("")
    g.P("   return r")
    g.P("}")
}

func (p *Generator) generateViewTemplate(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())

    g.P(`// RenderView will take in a http writer and object to render the view`)
    g.P(`func (x *`, typeName, `) RenderView(w http.ResponseWriter) {`)
    g.P("   tmpl, err := template.New(\"view\").Parse(` ")
    if len(message.Fields) > 0 {
        for _, field := range message.Fields {
         g.P(`<p class="w-16">`)
         g.P("  <span>", field.GoName, "</span>")
         g.P("  <span> {{ .", field.GoName, " }} </span>")
         g.P("</p>")
        }
    }
    g.P(    "`)")
    g.P("")
    g.P("   err := tmpl.Execute(w, x)")
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


