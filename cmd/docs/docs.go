package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gernest/bongo"
)

func main() {
	docsDir := "docs"
	app := bongo.New()
	if err := app.Run(docsDir); err != nil {
		log.Fatal(err)
	}

	docsSite := filepath.Join(docsDir, bongo.OutputDir)
	ghPages := filepath.Join(docsDir, "gh-pages", "bongo")

	err := filepath.Walk(docsSite, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out := strings.TrimPrefix(path, docsSite)
		dest := filepath.Join(ghPages, out)

		os.MkdirAll(filepath.Dir(dest), 0755)

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(dest, b, info.Mode())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
