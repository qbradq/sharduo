package data

import (
	"embed"
	"path"
)

// FS is the embedded file system
//
//go:embed templates lists
var FS embed.FS

// Walk walks the given path in the embedded file system
func Walk(dirPath string, fn func(string, []byte) []error) []error {
	var errs []error
	files, err := FS.ReadDir(dirPath)
	if err != nil {
		return []error{err}
	}
	for _, file := range files {
		if file.IsDir() {
			errs = append(errs, Walk(path.Join(dirPath, file.Name()), fn)...)
		} else {
			fpath := path.Join(dirPath, file.Name())
			d, err := FS.ReadFile(fpath)
			if err != nil {
				errs = append(errs, err)
			} else {
				errs = append(errs, fn(fpath, d)...)
			}
		}
	}
	return errs
}
