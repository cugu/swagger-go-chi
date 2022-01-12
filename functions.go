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
	"camel":             camel,
	"gotype":            gotype,
	"omitempty":         omitempty,
	"tojson":            tojson,
	"export":            camel,
	"parametertype":     parametertype,
	"parametertypename": parametertypename,
	"responsetype":      responsetype,
	"path":              fullpath,
	"body":              body,
	"dict":              dict,
	"roles":             roles,
}

func gotype(name string, s Schema, required []string) string {
	_, x := sgotype("", name, s, required, false)
	return x
}

func parametertype(parameter Parameter) string {
	reqs := []string{parameter.Name}
	if !parameter.Required {
		reqs = nil
	}

	s := parameter.Schema
	if parameter.Direct.Type != "" {
		s = parameter.Direct
	}

	_, x := sgotype("model.", parameter.Name, s, reqs, false)
	return x
}

func parametertypename(parameterType string) string {
	name := strcase.ToCamel(parameterType)
	if strings.HasPrefix(parameterType, "[]") {
		name += "Array"
	}

	return strings.Title(name)
}

func responsetype(responses map[string]*Response) string {
	response := responses["200"]

	_, x := sgotype("model.", "result", response.Schema, []string{"result"}, false)
	return x
}

func sgotype(pkg, name string, s Schema, required []string, nopointer bool) (bool, string) {
	req := ""
	if !nopointer && !contains(required, name) {
		req = "*"
	}

	if s.Ref != "" {
		return false, "*" + pkg + path.Base(s.Ref)
	}

	primitive := false
	t := ""

	switch s.Type {
	case "string":
		if s.Format == "date-time" {
			t = req + "time.Time"
		} else {
			t = req + "string"
			primitive = true
		}
	case "boolean":
		t = req + "bool"
		primitive = true
	case "object":
		if s.AdditionalProperties != nil {
			_, subType := sgotype(pkg, name, *s.AdditionalProperties, required, true)
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
		primitive = true
	case "array":
		_, subType := sgotype(pkg, name, *s.Items, required, true)
		t = "[]" + subType
	case "":
		t = "interface{}"
	default:
		panic(fmt.Sprintf("%#v", s))
	}

	return primitive, t
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

func tojson(name string, i Definition) string {
	b, _ := json.Marshal(i)
	b, _ = sjson.SetBytes(b, "$id", "#/definitions/"+name)
	return string(b)
}

func camel(s string) string {
	if s == "id" {
		return "ID"
	}
	return strings.Title(strcase.ToCamel(s))
}

func fullpath(lpath string, parameters []*Parameter) string {
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

func body(parameters []*Parameter) interface{} {
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
		var roles []string
		for _, scope := range req.Roles {
			roles = append(roles, scope)
			// roles = append(roles, permission.FromString(scope))
		}
		return strings.Join(roles, ", ")
	}
	return ""
}
