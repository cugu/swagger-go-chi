package main

import (
	"log"
	"os"
	"path"

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

	for name, file := range files {
		if err := os.MkdirAll(path.Join(config.Directory, path.Dir(name)), os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(path.Join(config.Directory, name), file.Data, file.Mode); err != nil {
			return err
		}
	}
	return nil
}
