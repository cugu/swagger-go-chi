package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"testing/fstest"
	"text/template"

	"github.com/rogpeppe/go-internal/modfile"
	"gopkg.in/yaml.v3"
)

//go:embed api/*
var api embed.FS

//go:embed time/*
var time embed.FS

//go:embed pointer/*
var pointer embed.FS

var packages = []fs.FS{api, time, pointer}

//go:embed templates/*
var templateFS embed.FS

var generations = []*generation{
	{"api", "server.go", template.Must(template.New("server.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/server.gotmpl"))},
	{"api", "test_api.go", template.Must(template.New("test_api.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/test_api.gotmpl"))},
	{"model", "model.go", template.Must(template.New("model.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/model.gotmpl"))},
	{"auth", "auth.go", template.Must(template.New("auth.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/auth.gotmpl"))},
	{"cli", "cli.go", template.Must(template.New("cli.gotmpl").Funcs(funcs).ParseFS(templateFS, "templates/cli.gotmpl"))},
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

	for _, fsys := range packages {
		err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			b, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			files[path] = &fstest.MapFile{Data: b, Mode: os.ModePerm}
			return nil
		})
		if err != nil {
			return nil, err
		}
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
