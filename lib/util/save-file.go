package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// SaveFileReader is the interface all save file implementations provide for
// reading saves.
type SaveFileReader interface {
	// Open opens the save file
	Open(p string) error
	// GetReader returns the io.ReadCloser for the named save file segment
	GetReader(segment string) io.ReadCloser
}

// SaveFileWriter is the interface all save file implementations provide for
// writing saves.
type SaveFileWriter interface {
	// Open opens the save file
	Open(p string) error
	// GetWriter returns the io.WriteCloser for the named save file segment.
	// This writer MUT BE CLOSED by the caller!
	GetWriter(segment string) io.WriteCloser
	// Close closes the save file. This MUST BE CALLED by the caller!
	Close() error
}

// CompressedSaveFileReader implements the SaveFile interface and creates a
// single zip file for the save.
type CompressedSaveFileReader struct {
	r *zip.ReadCloser
}

// Open implements the SaveFileReader interface
func (f *CompressedSaveFileReader) Open(p string) error {
	if info, err := os.Stat(); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("error opening %s: file does not exist", p)
		}
		if info.IsDir() {
			return fmt.Errorf("error opening %s: path is a directory", p)
		}
	}
	var err error
	f.r, err = zip.OpenReader(p)
	return err
}

// GetReader implements the SaveFileReader interface
func (f *CompressedSaveFileReader) GetReader(segment string) (io.ReadCloser, error) {
	return f.r.Open(segment)
}

// CompressedSaveFileWriter
