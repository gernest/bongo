package bongo

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/a8m/mark"
	"github.com/gernest/bongo/bindata/tpl"
	"github.com/gernest/front"
)

var (
	modTime           = "date"
	DefaultDateFormat = time.RFC822
	defaultView       = "post"

	defaultTpl = struct {
		home, index, page, post string
	}{
		"home.html",
		"index.html",
		"page.html",
		"post.html",
	}
)
var log = logrus.New()

type (
	// PageList is a collection of pages
	PageList []*Page

	// Page is a represantation of text document
	Page struct {
		Path string
		Body string
		HTML template.HTML
		Data interface{}
	}

	//FileLoader loads files needed for processing.
	// the filepaths can be relative or absolute.
	FileLoader func(string) ([]string, error)

	//FrontMatter extracts frontmatter from a text file
	FrontMatter interface {
		Parse(io.Reader) (front map[string]interface{}, body string, err error)
	}

	//Renderer renders the the projest.
	Renderer func(pages PageList, opts ...interface{}) error
)

// App is the main bongo appliaction.
type App struct {
	frontmatter FrontMatter
	fileLoader  FileLoader
	rendr       Renderer
	send        chan *Page
}

// NewApp creates a new App instance with default settings. You can overide the default
// behavior by implementing the interaces and use Set method to overide.
func NewApp() *App {
	matter := front.NewMatter()
	matter.Handle("---", front.YAMLHandler)
	return &App{
		fileLoader: func(dir string) ([]string, error) {

			// rst is the collection of files that are good for processing.
			// it only contains the filepaths( strings which allows access to the underlying
			// file.
			rst := []string{}

			// This checks if the file is a markdown file.
			// the check is based on file extension. It is assumed that file with
			// the following extensions are markdown files: .md, .MD, .markdown
			// and .mdown.
			var isMarkdown = func(file string) bool {
				ext := filepath.Ext(file)
				switch ext {
				case ".md", ".MD", ".markdown", ".mdown":
					return true
				}
				return false
			}

			// walkFunc walks the directory tree from path searching for files which
			// satisfy the markdown test( please see the var isMarkdown above). The
			// matching files are appended to the rst variable.
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

			err := filepath.Walk(dir, walkFunc) // lets walk
			return rst, err
		},
		frontmatter: matter,
		rendr: func(pgs PageList, opts ...interface{}) error {

			// loadTpl loads default templates. bongo default templates are embeded
			// by the go-bindata tool.
			//
			// So, this function  loads templates from embeded assets found in the
			// github.com/gernest/bongo/bindata/tpl package.
			//
			// If you want to see the command used to generate the tpl package please
			// see the bindata rule found in the Makefile at the root of this project.
			var loadTpl = func() (*template.Template, error) {
				t := template.New("bongo")
				tNames := []string{
					defaultTpl.home,
					defaultTpl.index,
					defaultTpl.page,
					defaultTpl.post,
				}
				for _, n := range tNames {
					tx := t.New(n)
					d, err := tpl.Asset(n)
					if err != nil {
						return nil, err
					}
					_, err = tx.Parse(string(d))
					if err != nil {
						return nil, err
					}
				}
				return t, nil
			}

			baseTpl, err := loadTpl() // load the templates.
			if err != nil {
				return fmt.Errorf("loading templates %v", err)
			}

			// render renders the page p using the template tmpl, and passing data as
			// template data context.
			//
			// The template used is the one loaded in the baseTpl variable.
			var render = func(p *Page, tmpl string, data interface{}) error {
				out := &bytes.Buffer{}
				rerr := baseTpl.ExecuteTemplate(out, tmpl, data)
				if rerr != nil {
					return rerr
				}
				p.HTML = template.HTML(out.String())
				return nil
			}

			trouble := make(chan error, 10) // For errors reporting
			done := make(chan bool, 10)     // If the process succeeded
			for k := range pgs {
				go func(tr chan error, good chan bool, id int) {
					page := pgs[id]
					var (
						view = defaultView
						data = make(map[string]interface{})
					)

					// we use the mark package to render markdown. The default configuration
					// for mark is prefered.
					//
					// I think there is no reason to keep the original markdown text. We only
					// keep the rendered text for further processing.
					page.Body = mark.New(page.Body, mark.DefaultOptions()).Render()
					switch page.Data.(type) {
					case map[string]interface{}:
						data = page.Data.(map[string]interface{})
						if v, ok := data["view"]; ok {
							view = v.(string)
						}
					}

					err := render(page, fmt.Sprintf("%s.html", view), data) // render the page
					if err != nil {
						tr <- fmt.Errorf("bongo: rendering %s %v", page.Path, err)
					}
					good <- true
				}(trouble, done, k)
			}

			// errs is a collection of errors accumulated in the rendering process above.
			var errs []error
			n := 0
		END:
			for {
				select {
				case err := <-trouble:
					errs = append(errs, err)
					n++
				case <-done:
					n++
				default:
					if len(pgs) <= n {
						break END
					}
				}
			}

			// We have to process the errs before returning it. To avoid implementing the
			// error interface.
			return func(args []error) error {
				if args != nil {
					rst := ""
					for k, v := range args {
						if k == 0 {
							rst = rst + v.Error()
						}
						rst = rst + ", \n" + v.Error()
					}
					return fmt.Errorf("%s", rst)
				}
				return nil
			}(errs)

		},
		send: make(chan *Page, 100),
	}
}

// Set overides default behaviour
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

// Run runs the app, and result in static content generation. This method is safe to run
// concurretly.
func (a *App) Run(root string) {
	files, err := a.fileLoader(root)
	if len(files) == 0 && err != nil {
		log.Fatalf("bongo: no files for processing %v", err)
	}
	var pages PageList
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
			if _, ok := front[modTime]; !ok {
				stat, err := f.Stat()
				if err != nil {
					log.Errorf("bongo: some fish adding timestamp to page %s %s", file, err)
					return
				}
				front[modTime] = stat.ModTime().Format(DefaultDateFormat)
			}
			a.send <- &Page{Path: file, Body: body, Data: front}
		}(f)
	}
	n := 0
END:
	for {
		select {
		case pg := <-a.send:
			pages = append(pages, pg)
			n++
		default:
			if len(files) <= n {
				break END
			}
		}

	}
	err = a.rendr(pages)
	if err != nil {
		log.Errorf("bongo: some fish rendering the project %v", err)
	}
}
