package bongo

import "path/filepath"

//HasExt hecks if the file has any mathing extension
func HasExt(file string, exts ...string) bool {
	fext := filepath.Ext(file)
	if len(exts) > 0 {
		for _, ext := range exts {
			if ext == fext {
				return true
			}
		}
	}
	return false
}
