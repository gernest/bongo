package bongo

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gernest/front"

	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

type (
	PageList []*Page

	Page struct {
		Path string
		Body string
		Data interface{}
	}

	FileLoader func(string) ([]string, error)

	FrontMatter interface {
		Parse(io.Reader) (front map[string]interface{}, body string, err error)
	}

	Renderer interface {
		Render(out io.Writer, tpl string, data interface{}) error
	}
)

type App struct {
	frontmatter FrontMatter
	fileLoader  FileLoader
	pages       PageList
	rendr       Renderer
	send        chan *Page
}

func NewApp() *App {
	matter := front.NewMatter()
	matter.Handle("---", front.YAMLHandler)
	return &App{
		fileLoader: func(dir string) ([]string, error) {
			rst := []string{}
			var isMarkdown = func(file string) bool {
				ext := filepath.Ext(file)
				switch ext {
				case ".md", ".MD", ".markdown", ".mdown":
					return true
				}
				return false
			}
			var walkFunc = func(path string, info os.FileInfo, err error) error {
				switch {
				case err != nil:
					return err
				case info.IsDir() || !isMarkdown(path):
					return nil
				}
				rst = append(rst, path)
				return nil
			}
			err := filepath.Walk(dir, walkFunc)
			return rst, err
		},
		frontmatter: matter,
		pages:       make(PageList, 0),
		send:        make(chan *Page, 100),
	}
}

func (a *App) Set(val interface{}) {
	switch val.(type) {
	case FrontMatter:
		a.frontmatter = val.(FrontMatter)
	case FileLoader:
		a.fileLoader = val.(FileLoader)
	case Renderer:
		a.rendr = val.(Renderer)
	}
}

func (a *App) Run(root string) {
	files, err := a.fileLoader(root)
	if len(files) == 0 && err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		go func(file string) {
			f, err := os.Open(file)
			if err != nil {
				log.Error(err)
				return
			}
			defer f.Close()
			front, body, err := a.frontmatter.Parse(f)
			if err != nil {
				log.Error(err)
				return
			}
			a.send <- &Page{Path: file, Body: body, Data: front}
		}(f)
	}
	n := 0
END:
	for {
		select {
		case pg := <-a.send:
			a.pages = append(a.pages, pg)
			n++
		default:
			if len(files) <= n {
				break END
			}
		}

	}
}
