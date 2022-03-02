package model

import (
	"time"

	"github.com/xeipuuv/gojsonschema"
)

var (
	schemaLoader = gojsonschema.NewSchemaLoader()
	UserSchema   = new(gojsonschema.Schema)
)

func init() {
	err := schemaLoader.AddSchemas(
		gojsonschema.NewStringLoader(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"$id":"#/definitions/User"}`),
	)
	if err != nil {
		panic(err)
	}

	UserSchema = mustCompile(`#/definitions/User`)
}

type User struct {
	Name string `json:"name"`
}

func mustCompile(uri string) *gojsonschema.Schema {
	s, err := schemaLoader.Compile(gojsonschema.NewReferenceLoader(uri))
	if err != nil {
		panic(err)
	}
	return s
}

const ()
