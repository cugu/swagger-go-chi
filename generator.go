package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"log"
	"os"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/rogpeppe/go-internal/modfile"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*
var templates embed.FS

type CLI struct {
	SwaggerYAML string `arg:"" name:"swagger" help:"Input swagger yaml" type:"existingfile"`
	Directory   string `arg:"" name:"path" help:"Destination path/package"`
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var cmd CLI
	kong.Parse(&cmd)

	f, err := os.Open(cmd.SwaggerYAML)
	if err != nil {
		return err
	}
	defer f.Close()

	swagger := &Swagger{}
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(swagger); err != nil {
		return err
	}

	b, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("error: %w, is swachigo run from the module root?", err)
	}

	swagger.Module = modfile.ModulePath(b)
	swagger.Package = cmd.Directory

	if err := generate(swagger, cmd.Directory, "model", "model"); err != nil {
		return err
	}
	if err := generate(swagger, cmd.Directory, "api", "api"); err != nil {
		return err
	}
	return generate(swagger, cmd.Directory, "api", "test_api")
}

func generate(s *Swagger, target, pkg, fileName string) error {
	t := template.New(fileName + ".gotmpl")
	t.Funcs(funcs)
	templ := template.Must(t.ParseFS(templates, "templates/"+fileName+".gotmpl"))

	if err := os.MkdirAll(target+"/"+pkg, os.ModePerm); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := templ.Execute(buf, s); err != nil {
		return err
	}

	fmtCode, err := format.Source(buf.Bytes())
	if err != nil {
		log.Println(err)
		fmtCode = buf.Bytes()
	}

	return os.WriteFile(target+"/"+pkg+"/"+fileName+".go", fmtCode, os.ModePerm)
}
