package bongo

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//Context contains the tree of processed pages.
type Context struct {
	Pages   PageList
	Tags    TagList
	Untaged PageList
	Data    map[string]interface{}
}

type Config struct {
	ThemeDir     string
	Theme        string
	DataDir      string
	TemplatesDir string
	OutputDir    string
}

//GetAllSections retruns a Context object for the PageList. Thismakes surepages
//are arranged by tags,  pages with no tags are assigned to the Cotext.Untagged.
func GetContext(p PageList) *Context {
	ctx := &Context{Pages: p}
	for i := 0; i < len(p); i++ {
		pg := p[i]
		for _, t := range pg.Tags {
			if len(ctx.Tags) > 0 {
				sort.Sort(ctx.Tags)
				key := sort.Search(len(ctx.Tags), func(x int) bool {
					return ctx.Tags[x].Name >= t
				})
				if key != len(ctx.Tags) {
					ctx.Tags[key].Pages = append(ctx.Tags[key].Pages, pg)
				}
				var pl PageList
				pl = append(pl, pg)
				ctx.Tags = append(ctx.Tags, &Tag{Name: t, Pages: pl})
			} else {
				ctx.Untaged = append(ctx.Untaged, pg)
			}
		}
	}
	return ctx
}

type ContextRender struct {
	ctx  *Context
	cfg  *Config
	tpl  *Template
	root string
}

func (r *ContextRender) Before(root string) error {
	configPath := filepath.Join(root, DefaultConfigFile)
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil
	}
	cfg := &Config{}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		return nil
	}
	r.cfg = cfg
	r.root = root
	tpl, err := LoadTemplates(root, cfg)
	if err != nil {
		return nil
	}
	r.tpl = tpl
	return nil
}

type Template struct {
	Theme *template.Template
	Main  *template.Template
}

func LoadTemplates(base string, cfg *Config) (*Template, error) {
	t := &Template{
		Theme: template.New("theme"),
		Main:  template.New("main"),
	}
	err := loadTpl(filepath.Join(base, cfg.ThemeDir, cfg.Theme), t.Theme)
	if err != nil {
		return nil, err
	}
	err = loadTpl(filepath.Join(base, cfg.TemplatesDir), t.Main)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func loadTpl(base string, tpl *template.Template) error {
	return filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !HasExt(path, defalutTplExtensions...) {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		n := strings.TrimPrefix(path, base)
		tx := tpl.New(n[1:])
		_, err = tx.Parse(string(b))
		if err != nil {
			return err
		}
		return nil
	})

}

func (t *Template) Lookup(name string) *template.Template {
	if theme := t.Theme.Lookup(name); theme != nil {
		return theme
	}
	return t.Main.Lookup(name)
}
