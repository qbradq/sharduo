package util

import (
	"archive/zip"
	"io"
	"os"
)

// ZipRead reads an entire zip file into memory and returns it as a dictionary
// of [io.Reader]s.
func ZipRead(path string) (map[string]io.ReadCloser, error) {
	ret := map[string]io.ReadCloser{}
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		ret[f.Name] = rc
	}
	return ret, nil
}

// ZipWrite writes all of the data in the readers in the dictionary and writes
// them to the zip using the key as the filename.
func ZipWrite(path string, d map[string]io.Reader) error {
	of, err := os.Create(path)
	if err != nil {
		return err
	}
	w := zip.NewWriter(of)
	for filename, r := range d {
		f, err := w.Create(filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}
	}
	return nil
}
