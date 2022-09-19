package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tidwall/sjson"
)

var funcs = map[string]interface{}{
	"goType":        goType,
	"parameterType": parameterType,
	"parameterName": parameterName,
	"responseType":  responseType,
	"omitempty":     omitempty,
	"toJSON":        toJSON,
	"export":        export,
	"examplePath":   examplePath,
	"exampleBody":   exampleBody,
	"dict":          dict,
	"roles":         roles,
}

func goType(name string, s *Schema, required []string) string {
	return schemaType("", name, s, required, false)
}

func parameterType(parameter Parameter) string {
	reqs := []string{parameter.Name}
	if !parameter.Required {
		reqs = nil
	}

	s := parameter.Schema
	if parameter.In == "formData" && parameter.Type != "file" {
		return "[]string" // This parameter type can be anything but parsing has to be done as string array
	}

	if parameter.Type != "" {
		s = &Schema{
			Format:      parameter.Format,
			Description: parameter.Description,
			Default:     parameter.Default,
			Maximum:     parameter.Maximum,
			Items:       parameter.Items,
			Type:        parameter.Type,
		}
	}

	return schemaType("model.", parameter.Name, s, reqs, false)
}

func parameterName(parameter Parameter) string {
	parameterType := parameterType(parameter)
	prefix := ""
	if strings.HasPrefix(parameterType, "*model.") {
		prefix = "model."
		parameterType = strings.TrimPrefix(parameterType, "*model.")
	}

	name := strcase.ToCamel(parameterType)
	if strings.HasPrefix(parameterType, "[]") {
		name += "Array"
	}

	return prefix + strings.Title(name)
}

func responseType(responses map[string]*Response) string {
	response := responses["200"]

	return schemaType("model.", "result", response.Schema, []string{"result"}, false)
}

func schemaType(pkg, name string, s *Schema, required []string, nopointer bool) string {
	req := ""
	if !nopointer && !contains(required, name) {
		req = "*"
	}

	if s.Ref != "" {
		return "*" + pkg + path.Base(s.Ref)
	}

	t := ""

	switch s.Type {
	case "string":
		if s.Format == "date-time" {
			t = req + "time.Time"
		} else {
			t = req + "string"
		}
	case "boolean":
		t = req + "bool"
	case "file":
		t = req + "[]*multipart.FileHeader"
	case "object":
		if s.AdditionalProperties != nil {
			subType := schemaType(pkg, name, s.AdditionalProperties, required, true)
			t = "map[string]" + subType
		} else {
			t = "map[string]interface{}"
		}
	case "number", "integer":
		if s.Format != "" {
			t = req + s.Format
		} else {
			t = req + "int"
		}
	case "array":
		subType := schemaType(pkg, name, s.Items, required, true)
		t = "[]" + subType
	case "":
		t = "interface{}"
	default:
		panic(fmt.Sprintf("%#v", s))
	}

	return t
}

func omitempty(name string, required []string) bool {
	return !contains(required, name)
}

func contains(required []string, name string) bool {
	for _, r := range required {
		if r == name {
			return true
		}
	}
	return false
}

func toJSON(name string, i Schema) string {
	b, _ := json.Marshal(i)
	b, _ = sjson.SetBytes(b, "$id", "#/definitions/"+name)
	return string(b)
}

func export(s string) string {
	if s == "id" {
		return "ID"
	}
	return strings.Title(strcase.ToCamel(s))
}

func examplePath(lpath string, parameters []*Parameter) string {
	u := url.URL{Path: lpath}
	q := u.Query()
	for _, p := range parameters {
		if p.In == "path" {
			if example := p.Examples; example != nil {
				u.Path = strings.ReplaceAll(u.Path, "{"+p.Name+"}", fmt.Sprint(example))
			}
		}
		if p.In == "query" {
			if example := p.Examples; example != nil {
				q.Set(p.Name, fmt.Sprint(example))
			}
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func exampleBody(parameters []*Parameter) interface{} {
	for _, p := range parameters {
		if p.In == "body" {
			if example := p.Examples; example != nil {
				return example
			}
		}
	}
	return nil
}

func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func roles(reqs []*Security) string {
	for _, req := range reqs {
		return strings.Join(req.Roles, ", ")
	}
	return ""
}
