package bongo

import (
	"os"

	"github.com/gernest/bongo-contrib/loaders"
	"github.com/gernest/bongo-contrib/matters"
	"github.com/gernest/bongo-contrib/models"
	"github.com/gernest/bongo-contrib/renderers"
)

type DefaultApp struct {
	loaders.DefaultLoader
	*matters.Matter
	renderers.DefaultRenderer
}

func newDefaultApp() *DefaultApp {
	app := &DefaultApp{}
	app.Matter = matters.NewYAML()
	return app
}

type App struct {
	gene models.Generator
}

func New() *App {
	return NewApp(newDefaultApp())
}

func NewApp(g models.Generator) *App {
	return &App{gene: g}
}

func (g *App) Run(root string) error {
	files, err := g.gene.Load(root)
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
			front, body, err := g.gene.Parse(f)
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
	err = g.gene.Render(root, pages)
	if err != nil {
		renderers.Rollback(root) // roll back before exiting
		return err
	}
	return nil

}
