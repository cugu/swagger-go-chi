package model

import (
	"time"

	"github.com/xeipuuv/gojsonschema"
)

var (
	schemaLoader    = gojsonschema.NewSchemaLoader()
	UserSchema      = new(gojsonschema.Schema)
	UserArraySchema = new(gojsonschema.Schema)
)

func init() {
	err := schemaLoader.AddSchemas(
		gojsonschema.NewStringLoader(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"$id":"#/definitions/User"}`),
		gojsonschema.NewStringLoader(`{"items":{"$ref":"#/definitions/User"},"type":"array","$id":"#/definitions/UserArray"}`),
	)
	if err != nil {
		panic(err)
	}

	UserSchema = mustCompile(`#/definitions/User`)
	UserArraySchema = mustCompile(`#/definitions/UserArray`)
}

type User struct {
	Name string `json:"name"`
}

type UserArray []*User

func mustCompile(uri string) *gojsonschema.Schema {
	s, err := schemaLoader.Compile(gojsonschema.NewReferenceLoader(uri))
	if err != nil {
		panic(err)
	}
	return s
}

const ()
