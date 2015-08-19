package bongo

import (
	"os"

	"github.com/gernest/bongo-contrib/loaders"
	"github.com/gernest/bongo-contrib/matters"
	"github.com/gernest/bongo-contrib/models"
	"github.com/gernest/bongo-contrib/renderers"
)

type Generator struct {
	loader models.FileLoader
	matter models.FrontMatter
	rendr  models.Renderer
}

func New() *Generator {
	return &Generator{
		loader: loaders.New(),
		matter: matters.NewYAML(),
		rendr:  renderers.New(),
	}
}

func (g *Generator) Run(root string) error {
	files, err := g.loader.Load(root)
	if err != nil {
		return err
	}
	var pages models.PageList
	send := make(chan *models.Page)
	errs := make(chan error)
	for _, f := range files {
		go func(file string) {
			f, err := os.Open(file)
			if err != nil {
				errs <- err
				return
			}
			defer f.Close()
			front, body, err := g.matter.Parse(f)
			if err != nil {
				errs <- err
				return
			}
			stat, err := f.Stat()
			if err != nil {
				errs <- err
				return
			}
			send <- &models.Page{Path: file, Body: body, Data: front, ModTime: stat.ModTime()}
		}(f)
	}
	n := 0
	var fish error
END:
	for {
		select {
		case pg := <-send:
			pages = append(pages, pg)
			n++
		case perr := <-errs:
			fish = perr
			break END

		default:
			if len(files) <= n {
				break END
			}
		}

	}
	if fish != nil {
		return fish
	}
	err = g.rendr.Render(root, pages)
	if err != nil {
		renderers.Rollback(root) // roll back before exiting
		return err
	}
	return nil

}
