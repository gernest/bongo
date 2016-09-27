package bongo

import (
	"html/template"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"github.com/a8m/mark"
)

const (

	//DefaultView is the default template view
	// for pages
	DefaultView = "post"

	//OutputDir is the name of the directory where generated files are saved
	OutputDir = "_site"

	//DefaultExt is the default extensinon name for output files
	DefaultExt = ".html"

	//DefaultPerm is the default permissinon for generated files
	DefaultPerm = 0600

	//DefaultPageKey is the key used to store current page in template
	// context data.
	DefaultPageKey = "Page"

	//CurrentSectionKey is the key used to store the current section value in the
	// template context
	CurrentSectionKey = "CurrentSection"

	//AllSectionsKey is the key used to store all sections in the template context data
	AllSectionsKey = "Sections"

	//DefaultConfigFile is the default configuraton file for abongo based project
	DefaultConfigFile = "_bongo.yml"

	//SiteConfigKey is the key used to store site wide configuration
	SiteConfigKey = "Site"

	//ThemeKey is the key used to store the name of the theme to be used
	ThemeKey = "theme"

	//ThemeDir is the directory where themes are installed
	ThemeDir = "_themes"

	//DefaultTheme the name of the default theme
	DefaultTheme = "gh"

	//StaticDir directory for static assets
	StaticDir = "static"

	cssDir         = "css"
	defaultSection = "home"
	pageSection    = "section"
	modTime        = "timeStamp"
)

//DefaultTpl is the defaut templates
var DefaultTpl = struct {
	Home, Index, Page, Post string
}{
	"home.html",
	"index.html",
	"page.html",
	"post.html",
}

type (
	// PageList is a collection of pages
	PageList []*Page

	// Page is a represantation of text document
	Page struct {
		Path    string
		Body    io.Reader
		ModTime time.Time
		Data    interface{}
		Tags    []string
	}

	//FileLoader loads files needed for processing.
	// the filepaths can be relative or absolute.
	FileLoader interface {
		Load(string) ([]string, error)
	}

	//FrontMatter extracts frontmatter from a text file
	FrontMatter interface {
		Parse(io.Reader) (front map[string]interface{}, body io.Reader, err error)
	}

	//Renderer generates static pages
	Renderer interface {
		Before(root string) error
		Render(root string, pages PageList, opts ...interface{}) error
		After(root string) error
	}

	//Generator is a static site generator
	Generator interface {
		FileLoader
		FrontMatter
		Renderer
	}
)

//HTML returns body text as html.
func (p *Page) HTML() template.HTML {
	b, _ := ioutil.ReadAll(p.Body)
	return template.HTML(mark.New(string(b), mark.DefaultOptions()).Render())
}

//
//
//	Sort Implementation for Pagelist
//
//

func (p PageList) Len() int           { return len(p) }
func (p PageList) Less(i, j int) bool { return p[i].ModTime.Before(p[j].ModTime) }
func (p PageList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// GetAllSections filter the pagelist for any section informations
// it returns a map of all the sections with the pages matching the
// section attached as a pagelist.
func GetAllSections(p PageList) map[string]PageList {
	sections := make(map[string]PageList)
	for k := range p {
		page := p[k]
		data := page.Data.(map[string]interface{})
		section := defaultSection

		if sec, ok := data[pageSection]; ok {
			switch sec.(type) {
			case string:
				section = sec.(string)
			}
		}
		if sdata, ok := sections[section]; ok {
			sdata = append(sdata, page)
			sections[section] = sdata
			continue
		}
		pList := make(PageList, 1)
		pList[0] = page
		sections[section] = pList
	}

	// sort the result before returning
	for key := range sections {
		sort.Sort(sections[key])
	}
	return sections
}
