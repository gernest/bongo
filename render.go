package bongo

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"github.com/gernest/gh"
	"gopkg.in/yaml.v2"
)

var (
	defaultTemplates     *template.Template
	defaultTheme         = "gh"
	defalutTplExtensions = []string{".html", ".tpl", ".tmpl"}
)

const (
	indexPage    = "index.html"
	statusPassed = iota
	statusFailed
	statusComplete
)

type buildStatus struct {
	Status  int
	Message string
}

func newStatus(code int, msg string) *buildStatus {
	return &buildStatus{Status: code, Message: msg}
}

func init() {
	defaultTemplates = template.New("bongo")
	for _, n := range gh.AssetNames() {
		if filepath.Ext(n) != ".html" {
			continue
		}
		tx := defaultTemplates.New(filepath.Join("gh", n))
		d, err := gh.Asset(n)
		if err != nil {
			panic(err)
		}
		_, err = tx.Parse(string(d))
		if err != nil {
			panic(err)
		}
	}
	log.SetFlags(log.Lshortfile)
}

//DefaultRenderer is the default REnderer implementation
type DefaultRenderer struct {
	config map[string]interface{}
	rendr  *template.Template
	root   string
}

// Before loads configurations and prepare rendering stuffs
func (d *DefaultRenderer) Before(root string) error {
	cfg, rendr, err := load(root)
	if err != nil {
		return err
	}
	d.config = cfg
	d.rendr = rendr
	d.root = root
	return nil
}

// Render builds a static site
func (d *DefaultRenderer) Render(root string, pages PageList, opts ...interface{}) error {
	baseInfo, err := os.Stat(root)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}

	buildDIr := filepath.Join(root, OutputDir)
	if err = prepareBuild(baseInfo, buildDIr); err != nil {
		return err
	}
	themeName := d.getTheme()

	allsections := GetAllSections(pages)
	for key := range allsections {
		setionPages := allsections[key]
		view := DefaultView
		data := make(map[string]interface{})

		data[CurrentSectionKey] = setionPages
		data[AllSectionsKey] = allsections

		for _, page := range setionPages {
			switch page.Data.(type) {
			case map[string]interface{}:
				if v, ok := page.Data.(map[string]interface{})[DefaultView]; ok {
					view = v.(string)
				}
			}
			data[DefaultPageKey] = page

			buf.Reset()
			rerr := d.rendr.ExecuteTemplate(buf, fmt.Sprintf("%s/%s.html", themeName, view), data)
			if rerr != nil {
				break
			}

			destFileName := strings.Replace(filepath.Base(page.Path), filepath.Ext(page.Path), DefaultExt, -1)
			destDir := filepath.Join(buildDIr, key)
			os.MkdirAll(destDir, baseInfo.Mode())
			destFile := filepath.Join(destDir, destFileName)

			ioerr := ioutil.WriteFile(destFile, buf.Bytes(), DefaultPerm)
			if ioerr != nil {
				break
			}

		}

		// write the index page for the section.
		buf.Reset()

		rerr := d.rendr.ExecuteTemplate(buf, filepath.Join(themeName, DefaultTpl.Index), data)
		if rerr != nil {
			return rerr
		}
		os.MkdirAll(filepath.Join(buildDIr, key), baseInfo.Mode())

		destIndexFile := filepath.Join(buildDIr, key, indexPage)
		ioerr := ioutil.WriteFile(destIndexFile, buf.Bytes(), DefaultPerm)
		if ioerr != nil {
			return ioerr
		}

	}

	// write home page.
	buf.Reset()

	data := make(map[string]interface{})
	data[AllSectionsKey] = allsections
	data[SiteConfigKey] = d.config

	rerr := d.rendr.ExecuteTemplate(buf, filepath.Join(themeName, DefaultTpl.Home), data)
	if rerr != nil {
		return rerr
	}

	homePage := filepath.Join(buildDIr, indexPage)
	ioerr := ioutil.WriteFile(homePage, buf.Bytes(), DefaultPerm)
	if ioerr != nil {
		return ioerr
	}
	return nil
}

func (d *DefaultRenderer) getTheme() string {
	return d.config[ThemeKey].(string)
}

// After copies relevant static files to the generated site
func (d *DefaultRenderer) After(root string) error {
	return d.copyStatic()
}

func (d *DefaultRenderer) copyStatic() error {
	theme := d.getTheme()
	out := d.getOutputDir()
	switch theme {
	case defaultTheme:
		info, _ := os.Stat(d.root)
		for _, f := range gh.AssetNames() {
			if filepath.Ext(f) == ".html" {
				continue
			}
			base := filepath.Join(out, filepath.Dir(f))
			os.MkdirAll(base, info.Mode())
			b, err := gh.Asset(f)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(filepath.Join(out, f), b, DefaultPerm)
			if err != nil {
				log.Println(err, base)
				return err
			}
		}
		return nil
	default:
		// we copy the static directory in the current theme directory
		staticDir := filepath.Join(d.root, ThemeDir, theme, StaticDir)
		if com.IsExist(staticDir) {
			err := com.CopyDir(staticDir, filepath.Join(d.root, OutputDir, StaticDir))
			if err != nil {
				return err
			}
		}

	}

	// if we have the static set on the config file we use it.
	if staticDirective, ok := d.config[StaticDir]; ok {
		switch staticDirective.(type) {
		case []interface{}:
			sd := staticDirective.([]interface{})
			for _, f := range sd {
				dir := f.(string)
				err := com.CopyDir(filepath.Join(d.root, dir), filepath.Join(out, dir))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *DefaultRenderer) getOutputDir() string {
	return filepath.Join(d.root, OutputDir)
}

//NewDefaultRenderer returns default Renderer implementation
func NewDefaultRenderer() *DefaultRenderer {
	return &DefaultRenderer{config: make(map[string]interface{})}
}

func prepareBuild(baseInfo os.FileInfo, buildDir string) error {
	// If there is already a built project we remove it and start afresh
	info, err := os.Stat(buildDir)
	if err != nil {
		if os.IsNotExist(err) {
			oerr := os.MkdirAll(buildDir, baseInfo.Mode())
			if oerr != nil {
				return fmt.Errorf("create build dir at %s %v", buildDir, err)
			}
		}
	} else {
		oerr := os.RemoveAll(buildDir)
		if oerr != nil {
			return fmt.Errorf("cleaning %s %v", buildDir, oerr)
		}
		oerr = os.MkdirAll(buildDir, info.Mode())
		if oerr != nil {
			return fmt.Errorf("creating %s %v", buildDir, oerr)
		}
	}
	return nil
}

//Rollback delets the build directory
func Rollback(root string) {
	buildDIr := filepath.Join(root, OutputDir)
	os.RemoveAll(buildDIr)
}

func loadTheme(base, name string, tpl *template.Template) error {
	themesDir := filepath.Join(base, ThemeDir)
	return filepath.Walk(filepath.Join(themesDir, name), func(path string, info os.FileInfo, err error) error {
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
		n := strings.TrimPrefix(path, themesDir)
		tx := tpl.New(n[1:])
		_, err = tx.Parse(string(b))
		if err != nil {
			return err
		}
		return nil
	})

}

func load(root string) (map[string]interface{}, *template.Template, error) {
	cfg := loadConfig(root)
	tpl := template.New("bongo")
	if cfg != nil {
		if tName, ok := cfg[ThemeKey]; ok {
			terr := loadTheme(root, tName.(string), tpl)
			if terr != nil {
				return nil, nil, terr
			}
			return cfg, tpl, nil
		}
		cfg[ThemeKey] = defaultTheme
		return cfg, defaultTemplates, nil
	}
	c := make(map[string]interface{})
	c[ThemeKey] = defaultTheme
	return c, defaultTemplates, nil
}

func loadConfig(root string) map[string]interface{} {
	configPath := filepath.Join(root, DefaultConfigFile)
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal(b, m)
	if err != nil {
		log.Fatalf("loading config %v \n", err)
	}
	return m
}
