package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_exampleBody(t *testing.T) {
	type args struct {
		parameters []*Parameter
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{"simple", args{[]*Parameter{{Name: "user", In: "body", Required: true, Schema: &Schema{Ref: "#/definitions/User"}, Examples: map[string]interface{}{"name": "bob"}}}}, map[string]interface{}{"name": "bob"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, exampleBody(tt.args.parameters), "body(%v)", tt.args.parameters)
		})
	}
}

func Test_contains(t *testing.T) {
	type args struct {
		required []string
		name     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"element exists", args{[]string{"a", "b"}, "a"}, true},
		{"element does not exist", args{[]string{"a", "b"}, "c"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, contains(tt.args.required, tt.args.name), "contains(%v, %v)", tt.args.required, tt.args.name)
		})
	}
}

func Test_dict(t *testing.T) {
	type args struct {
		values []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{"simple", args{[]interface{}{"a", "b"}}, map[string]interface{}{"a": "b"}, func(assert.TestingT, error, ...interface{}) bool { return false }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dict(tt.args.values...)
			if !tt.wantErr(t, err, fmt.Sprintf("dict(%v)", tt.args.values...)) {
				return
			}
			assert.Equalf(t, tt.want, got, "dict(%v)", tt.args.values...)
		})
	}
}

func Test_export(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"hello"}, "Hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, export(tt.args.s), "export(%v)", tt.args.s)
		})
	}
}

func Test_examplePath(t *testing.T) {
	type args struct {
		lpath      string
		parameters []*Parameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"/user/{id}", []*Parameter{{Name: "id", In: "path", Required: true, Type: "string", Examples: "123"}}}, "/user/123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, examplePath(tt.args.lpath, tt.args.parameters), "fullpath(%v, %v)", tt.args.lpath, tt.args.parameters)
		})
	}
}

func Test_goType(t *testing.T) {
	type args struct {
		name     string
		s        *Schema
		required []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"string", args{"id", &Schema{Type: "string"}, []string{"id"}}, "string"},
		{"optional string", args{"id", &Schema{Type: "string"}, []string{}}, "*string"},
		{"custom type", args{"user", &Schema{Ref: "#/definitions/User"}, []string{"user"}}, "*User"},
		{"optional custom type", args{"user", &Schema{Ref: "#/definitions/User"}, []string{}}, "*User"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, goType(tt.args.name, tt.args.s, tt.args.required), "gotype(%v, %v, %v)", tt.args.name, tt.args.s, tt.args.required)
		})
	}
}

func Test_omitempty(t *testing.T) {
	type args struct {
		name     string
		required []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"do not omit", args{"id", []string{"id"}}, false},
		{"omitempty", args{"id", []string{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, omitempty(tt.args.name, tt.args.required), "omitempty(%v, %v)", tt.args.name, tt.args.required)
		})
	}
}

func Test_parameterType(t *testing.T) {
	type args struct {
		parameter Parameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{Parameter{Name: "id", In: "path", Required: true, Type: "string", Examples: "123"}}, "string"},
		{"optional string", args{Parameter{Name: "id", In: "query", Required: false, Type: "string", Examples: "123"}}, "*string"},
		{"optional integer", args{Parameter{Name: "no", In: "query", Required: false, Type: "integer", Format: "int64", Examples: 123}}, "*int64"},
		{"custom", args{Parameter{Name: "user", In: "body", Required: true, Schema: &Schema{Ref: "#/definitions/User"}}}, "*model.User"},
		{"custom array", args{Parameter{Name: "users", In: "body", Required: true, Schema: &Schema{Type: "array", Items: &Schema{Ref: "#/definitions/User"}}}}, "[]*model.User"},
		{"formData string", args{Parameter{Name: "users", In: "formData", Required: true, Type: "string"}}, "[]string"},
		{"formData file", args{Parameter{Name: "users", In: "formData", Required: true, Type: "file"}}, "[]*multipart.FileHeader"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parameterType(tt.args.parameter), "parametertype(%v)", tt.args.parameter)
		})
	}
}

func Test_parameterName(t *testing.T) {
	type args struct {
		parameter Parameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"string", args{Parameter{Name: "id", In: "path", Required: true, Type: "string"}}, "String"},
		{"string pointer", args{Parameter{Name: "id", In: "query", Required: false, Type: "string"}}, "String"},
		{"string array", args{Parameter{Name: "id", In: "query", Required: false, Type: "array", Items: &Schema{Type: "string"}}}, "StringArray"},
		// {"custom", args{Parameter{Name: "user", In: "body", Required: true, Schema: Schema{Ref: "#/definitions/User"}}}, "xString"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parameterName(tt.args.parameter), "parametertypename(%v)", tt.args.parameter)
		})
	}
}

func Test_responseType(t *testing.T) {
	type args struct {
		responses map[string]*Response
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{map[string]*Response{"200": {Schema: &Schema{Type: "string"}}}}, "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, responseType(tt.args.responses), "responsetype(%v)", tt.args.responses)
		})
	}
}

func Test_roles(t *testing.T) {
	type args struct {
		reqs []*Security
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{[]*Security{{Roles: []string{"admin"}}}}, "admin"},
		{"list", args{[]*Security{{Roles: []string{"user", "foo"}}}}, "user, foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, roles(tt.args.reqs), "roles(%v)", tt.args.reqs)
		})
	}
}

func Test_schemaType(t *testing.T) {
	type args struct {
		pkg       string
		name      string
		s         *Schema
		required  []string
		nopointer bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"string", args{"model.", "result", &Schema{Type: "string"}, []string{"result"}, false}, "string"},
		{"custom", args{"model.", "result", &Schema{Ref: "#/definitions/User"}, []string{"result"}, false}, "*model.User"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := schemaType(tt.args.pkg, tt.args.name, tt.args.s, tt.args.required, tt.args.nopointer)
			assert.Equalf(t, tt.want, got1, "sgotype(%v, %v, %v, %v, %v)", tt.args.pkg, tt.args.name, tt.args.s, tt.args.required, tt.args.nopointer)
		})
	}
}

func Test_toJSON(t *testing.T) {
	type args struct {
		name string
		i    Schema
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"Name", Schema{Type: "string"}}, "{\"type\":\"string\",\"$id\":\"#/definitions/Name\"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, toJSON(tt.args.name, tt.args.i), "tojson(%v, %v)", tt.args.name, tt.args.i)
		})
	}
}
