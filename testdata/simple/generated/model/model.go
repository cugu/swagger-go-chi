package model

import (
	"time"

	"github.com/xeipuuv/gojsonschema"
)

var (
	schemaLoader = gojsonschema.NewSchemaLoader()
)

func init() {
	err := schemaLoader.AddSchemas()
	if err != nil {
		panic(err)
	}

}

func mustCompile(uri string) *gojsonschema.Schema {
	s, err := schemaLoader.Compile(gojsonschema.NewReferenceLoader(uri))
	if err != nil {
		panic(err)
	}
	return s
}

const ()
