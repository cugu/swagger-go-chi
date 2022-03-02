package main

import (
	"log"
	"os"
	"path"
	"testing/fstest"

	"github.com/alecthomas/kong"
)

type Config struct {
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

	config := &Config{}
	kong.Parse(config)

	yamlData, err := os.ReadFile(config.SwaggerYAML)
	if err != nil {
		return err
	}

	modulePath, err := modulePath()
	if err != nil {
		return err
	}

	files, err := generate(modulePath+"/"+config.Directory, yamlData)
	if err != nil {
		return err
	}

	return writeFiles(files, config.Directory)
}

func writeFiles(files fstest.MapFS, dst string) error {
	for name, file := range files {
		if err := os.MkdirAll(path.Join(dst, path.Dir(name)), os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(path.Join(dst, name), file.Data, file.Mode); err != nil {
			return err
		}
	}
	return nil
}
