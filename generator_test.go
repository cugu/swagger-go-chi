package main

import (
	"io"
	"io/fs"
	"os"
	"path"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func Test_generate(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range entries {
		t.Run(dir.Name(), func(t *testing.T) {
			yamlData, err := os.ReadFile(path.Join("testdata", dir.Name(), "swagger.yml"))
			if err != nil {
				t.Fatal(err)
			}

			got, err := generate("github.com/cugu/swagger-go-chi/testdata/"+dir.Name()+"/generated", yamlData)
			if (err != nil) != false {
				t.Errorf("generate() error = %v, wantErr %v", err, false)
				return
			}

			want, err := toMapFS(path.Join("testdata", dir.Name(), "generated"))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, want, got)
			assertFS(t, want, got)
		})
	}
}

func assertFS(t *testing.T, want, got fstest.MapFS) {
	wantNames := keys(want)
	gotNames := keys(got)
	assert.ElementsMatch(t, wantNames, gotNames)

	for _, name := range wantNames {
		assert.Equal(t, string(want[name].Data), string(got[name].Data))
	}
}

func keys(mapfs fstest.MapFS) []string {
	var names []string
	for name := range mapfs {
		names = append(names, name)
	}
	return names
}

func toMapFS(dirPath string) (fstest.MapFS, error) {
	want := fstest.MapFS{}
	fsys := os.DirFS(dirPath)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}

		f, err := fsys.Open(path)
		if err != nil {
			return err
		}

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		want[path] = &fstest.MapFile{Data: b, Mode: os.ModePerm}

		return nil
	})
	return want, err
}
