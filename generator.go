package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/rogpeppe/go-internal/modfile"
	"go/format"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"testing/fstest"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

var generations = []*generation{
	{"api", "api.go", template.Must(template.New("api.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/api.gotmpl"))},
	{"api", "test_api.go", template.Must(template.New("test_api.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/test_api.gotmpl"))},
	{"model", "model.go", template.Must(template.New("model.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/model.gotmpl"))},
}

type generation struct {
	Package  string
	Name     string
	Template *template.Template
}

type TemplateData struct {
	ImportPath string
	Swagger    *Swagger
}

func generate(importPath string, yamlData []byte) (fstest.MapFS, error) {
	swagger := &Swagger{}
	if err := yaml.Unmarshal(yamlData, swagger); err != nil {
		return nil, err
	}

	data := &TemplateData{
		ImportPath: importPath,
		Swagger:    swagger,
	}

	files := fstest.MapFS{}

	for _, templ := range generations {
		buf := &bytes.Buffer{}
		if err := templ.Template.Execute(buf, data); err != nil {
			return nil, err
		}

		fmtCode, err := format.Source(buf.Bytes())
		if err != nil {
			log.Println(err)
			fmtCode = buf.Bytes()
		}
		files[templ.Package+"/"+templ.Name] = &fstest.MapFile{Data: fmtCode, Mode: os.ModePerm}
	}

	return files, nil
}

func modulePath() (string, error) {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("error: %w, is swachigo run from the module root?", err)
	}
	return modfile.ModulePath(b), nil
}
