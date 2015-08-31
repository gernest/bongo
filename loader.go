package bongo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bongo-contrib/models"

	"github.com/bongo-contrib/utils"
)

var supportedExtensions = []string{".md", ".MD", "..markdown"}

//DefaultLoader  is the default FileLoader implementation
type DefaultLoader struct{}

// NewLoader returns default FileLoader implementation.
func NewLoader() *DefaultLoader {
	return &DefaultLoader{}
}

// Load loads files found in the base path for processing.
func (d DefaultLoader) Load(base string) ([]string, error) {
	return func(base string) ([]string, error) {
		var rst []string
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			switch {
			case err != nil:
				return err
			case info.IsDir():
				return nil
			case !utils.HasExt(path, supportedExtensions...):
				return nil
			case strings.Contains(path, models.OutputDir):
				return nil
			}
			rst = append(rst, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return rst, nil

	}(base)

}
