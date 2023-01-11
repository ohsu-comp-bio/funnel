package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

type Field struct {
	Name     string
	Type     string
	Repeated bool
	ID       int
	Source   *openapi3.SchemaRef
}

type Message struct {
	Name   string
	Fields []Field
}

type Enum struct {
	Name   string
	Values []interface{}
}

type ServicePath struct {
	Name       string
	Path       string
	Mode       string
	InputType  string
	OutputType string
}

func getType(p *openapi3.SchemaRef) (bool, string) {
	switch p.Value.Type {
	case "integer":
		return false, "int32"
	case "boolean":
		return false, "bool"
	case "number":
		return false, "double"
	case "object":
		if p.Ref != "" {
			t := strings.Split(p.Ref, "/")
			return false, t[len(t)-1]
		}
		if p.Value.AdditionalProperties != nil {
			_, aType := getType(p.Value.AdditionalProperties)
			return false, fmt.Sprintf("map<string,%s>", aType)
		}
		return false, "map<string,string>"
		//return fmt.Sprintf("%#v", p.Value)
	case "array":
		if p.Value.Items.Ref != "" {
			t := strings.Split(p.Value.Items.Ref, "/")
			return true, t[len(t)-1]
		}
		_, aType := getType(p.Value.Items)
		return true, aType
	default:
		if p.Ref != "" {
			t := strings.Split(p.Ref, "/")
			return false, t[len(t)-1]
		}
		return false, p.Value.Type
	}
}

func getParamType(param *openapi3.Parameter) (bool, string) {
	//log.Printf("Param %#v\n", param)

	if param.Schema.Ref != "" {
		t := strings.Split(param.Schema.Ref, "/")
		return false, t[len(t)-1]
	}

	if param.Schema.Value.Type != "" {
		return getType(param.Schema)
	}
	return false, ""

}

func getFields(schema *openapi3.SchemaRef) []Field {
	fields := []Field{}
	for pname, pschema := range schema.Value.Properties {
		//	fmt.Printf("\t%s %#v\n", pname, pschema)
		repeated, fType := getType(pschema)
		f := Field{Name: pname, Repeated: repeated, Type: fType, Source: schema}
		fields = append(fields, f)
	}
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	for i := range fields {
		fields[i].ID = i + 1
	}
	return fields
}

func parseMessageSchema(name string, schema *openapi3.SchemaRef) (Message, error) {
	if schema.Value.Properties != nil {
		//fmt.Printf("%s %#v\n", name, schema.Value.Properties)
		fields := getFields(schema)
		m := Message{Name: name, Fields: fields}
		return m, nil
	} else if schema.Value.AllOf != nil {
		//fmt.Printf("All of %s %#v\n", name, schema.Value.AllOf)
		fieldMap := map[string]Field{}
		for i := range schema.Value.AllOf {
			newFields := getFields(schema.Value.AllOf[i])
			//fmt.Printf("\t%#v\n", newFields)
			for _, f := range newFields {
				if x, ok := fieldMap[f.Name]; !ok {
					fieldMap[f.Name] = f
				} else {
					//if same field from two sources, pick local one
					if x.Source.Ref != "" && f.Source.Ref == "" {
						fieldMap[f.Name] = f
					}
				}
			}
		}
		fields := []Field{}
		for _, v := range fieldMap {
			fields = append(fields, v)
		}
		//fmt.Printf("Fields: %#vs\n", fields)
		sort.SliceStable(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		for i := range fields {
			fields[i].ID = i + 1
		}
		m := Message{Name: name, Fields: fields}
		return m, nil
	} else if schema.Value.Enum != nil {
		return Message{}, fmt.Errorf("Message is Enum")
	} else {
		m := Message{Name: name, Fields: []Field{}}
		return m, nil
	}
}

func parseMessageEnum(name string, schema *openapi3.SchemaRef) (Enum, error) {
	e := Enum{Name: name, Values: schema.Value.Enum}
	return e, nil
}

func cleanSchema(messages []Message, enums []Enum, services []ServicePath) {
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].Name < messages[j].Name })

	var prefixRemove = "tes"
	for i := range messages {
		messages[i].Name = strings.TrimPrefix(messages[i].Name, prefixRemove)
		for j := range messages[i].Fields {
			messages[i].Fields[j].Type = strings.TrimPrefix(messages[i].Fields[j].Type, prefixRemove)
		}
	}

	for i := range enums {
		enums[i].Name = strings.TrimPrefix(enums[i].Name, prefixRemove)
	}

	for i := range services {
		services[i].InputType = strings.TrimPrefix(services[i].InputType, prefixRemove)
		services[i].OutputType = strings.TrimPrefix(services[i].OutputType, prefixRemove)
	}
}

func getResponseMessage(resp *openapi3.Responses) string {
	schema := resp.Get(200).Value.Content.Get("application/json").Schema
	s := strings.Split(schema.Ref, "/")
	return s[len(s)-1]
}

func main() {
	flag.Parse()

	input := flag.Arg(0)

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(input)

	messages := []Message{}
	enums := []Enum{}

	service := []ServicePath{}

	if err != nil {
		log.Printf("Parsing Error: %s\n", err)
	} else {
		//fmt.Printf("%#v\n", doc.Components.Parameters)
		for name, schema := range doc.Components.Schemas {
			if schema.Value.Enum != nil {
				if e, err := parseMessageEnum(name, schema); err == nil {
					enums = append(enums, e)
				}
			} else {
				if e, err := parseMessageSchema(name, schema); err == nil {
					messages = append(messages, e)
				}
			}
		}

		for name, param := range doc.Components.Parameters {
			//fmt.Printf("component params: %s %#v\n", name, param.Value.Schema.Value)
			if param.Value.Schema.Value.Enum != nil {
				if e, err := parseMessageEnum(name, param.Value.Schema); err == nil {
					//fmt.Printf("param %s %#v\n", name, e)
					enums = append(enums, e)
				}
			}
		}

		for path, req := range doc.Paths {
			if req.Get != nil {
				//log.Printf("Get: %s %s\n", path, req.Get.OperationID)
				reqFields := []Field{}
				for _, param := range req.Get.Parameters {
					r, t := getParamType(param.Value)
					reqFields = append(reqFields, Field{Name: param.Value.Name, Type: t, Repeated: r})
				}
				for i := range reqFields {
					reqFields[i].ID = i + 1
				}
				m := Message{Name: req.Get.OperationID + "Request", Fields: reqFields}
				messages = append(messages, m)

			}
			if req.Post != nil {
				log.Printf("Post: %s %s\n", path, req.Post.OperationID)
				reqFields := []Field{}
				for _, param := range req.Post.Parameters {
					r, t := getParamType(param.Value)
					reqFields = append(reqFields, Field{Name: param.Value.Name, Type: t, Repeated: r})
				}
				for i := range reqFields {
					reqFields[i].ID = i + 1
				}
				if req.Post.RequestBody != nil {
					//if not a reference to schema, build it
				} else {
					m := Message{Name: req.Post.OperationID + "Request", Fields: reqFields}
					messages = append(messages, m)
				}
			}
		}

		for path, req := range doc.Paths {
			p := ServicePath{}
			if req.Get != nil {
				p.Name = req.Get.OperationID
				p.Path = path
				p.Mode = "get"
				p.InputType = req.Get.OperationID + "Request"
				p.OutputType = getResponseMessage(&req.Get.Responses)
				service = append(service, p)
			}
			if req.Post != nil {
				p.Name = req.Post.OperationID
				p.Path = path
				p.Mode = "post"
				if req.Post.RequestBody != nil {
					s := req.Post.RequestBody.Value.Content.Get("application/json").Schema.Ref
					sL := strings.Split(s, "/")
					p.InputType = sL[len(sL)-1]
					log.Printf("post %s ref %s", path, p.InputType)
				} else {
					p.InputType = req.Post.OperationID + "Request"
				}
				p.OutputType = getResponseMessage(&req.Post.Responses)
				service = append(service, p)
			}
		}

	}

	cleanSchema(messages, enums, service)

	tmpl, err := template.New("proto").Parse(`
syntax = "proto3";

option go_package = "github.com/ohsu-comp-bio/funnel/tes";

package tes;

import "google/api/annotations.proto";

{{range $i, $enum := .enums}}
enum {{$enum.Name}} { {{range $j, $value := $enum.Values}}
	{{$value}} = {{$j}};{{end}}
}
{{end}}
{{range $i, $message := .messages}}
message {{$message.Name}} { {{range $j, $field := $message.Fields}}
	{{if $field.Repeated}}repeated {{end}}{{$field.Type}} {{$field.Name}} = {{$field.ID}};{{end}}
}
{{end}}

service TaskService {
{{range $i, $path := .services}}
    rpc {{$path.Name}}({{$path.InputType}}) returns ({{$path.OutputType}}) {
      option (google.api.http) = {
        {{$path.Mode}}: "{{$path.Path}}"
		additional_bindings {
			{{$path.Mode}}: "/v1{{$path.Path}}"
			{{- if eq $path.Mode "post"}}
			body: "*"{{end}}
		}
		additional_bindings {
			{{$path.Mode}}: "/ga4gh/tes/v1{{$path.Path}}"
			{{- if eq $path.Mode "post"}}
			body: "*"{{end}}
		}
		{{- if eq $path.Mode "post"}}
		body: "*"{{end}}
      };
    }
{{end}}
}

`)

	_ = tmpl
	if err != nil {
		fmt.Printf("Template Error: %s\n", err)
	} else {
		err := tmpl.Execute(os.Stdout, map[string]interface{}{"messages": messages, "enums": enums, "services": service})
		if err != nil {
			log.Fatalf("Template Error: %s", err)
		}
	}
}
