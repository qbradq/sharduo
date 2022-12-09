package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// SaveFileReader is the interface all save file implementations provide for
// reading saves.
type SaveFileReader interface {
	// Open opens the save file
	Open(p string) error
	// GetReader returns the io.ReadCloser for the named save file segment.
	// Close MUST BE CALLED by the caller!
	GetReader(segment string) (io.ReadCloser, error)
	// Close closes the save file. This MUST BE CALLED by the caller!
	Close() error
}

// SaveFileWriter is the interface all save file implementations provide for
// writing saves.
type SaveFileWriter interface {
	// Open opens the save file
	Open(p string) error
	// GetWriter returns the io.WriteCloser for the named save file segment.
	// Close MUST BE CALLED by the caller!
	GetWriter(segment string) (io.WriteCloser, error)
	// Close closes the save file. This MUST BE CALLED by the caller!
	Close() error
}

// CompressedSaveFileReader implements the SaveFileReader interface and reads a
// single zip file for the save.
type CompressedSaveFileReader struct {
	r *zip.ReadCloser
}

// NewCompressedSaveFileReader returns a new CompressedSaveFileReader object
func NewCompressedSaveFileReader() *CompressedSaveFileReader {
	return &CompressedSaveFileReader{}
}

// Open implements the SaveFileReader interface
func (f *CompressedSaveFileReader) Open(p string) error {
	if info, err := os.Stat(p); err != nil {
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

// Close implements the SaveFileReader interface
func (f *CompressedSaveFileReader) Close() error {
	return nil
}

// CompressedSaveFileWriter implements the SaveFileWriter interface.
type CompressedSaveFileWriter struct {
	w *zip.Writer
}

// NewCompressedSaveFileReader returns a new CompressedSaveFileReader object
func NewCompressedSaveFileWriter() *CompressedSaveFileWriter {
	return &CompressedSaveFileWriter{}
}

// Open implements the SaveFileWriter interface.
func (f *CompressedSaveFileWriter) Open(p string) error {
	fullPath := p
	if filepath.Ext(p) != "zip" {
		fullPath += ".zip"
	}
	fullPath = filepath.Clean(fullPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0777); err != nil {
		return err
	}
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	f.w = zip.NewWriter(file)
	return nil
}

// GetWriter implements the SaveFileWriter interface.
func (f *CompressedSaveFileWriter) GetWriter(segment string) (io.Writer, error) {
	o, err := f.w.Create(segment)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Close implements the SaveFileWriter interface.
func (f *CompressedSaveFileWriter) Close() error {
	return f.w.Close()
}

// DebugSaveFileReader reads data sets from individual files from a directory.
type DebugSaveFileReader struct {
	// Path to the directory for the save files
	savePath string
}

// NewDebugSaveFileReader creates a new DebugSaveFileReader and returns it.
func NewDebugSaveFileReader() *DebugSaveFileReader {
	return &DebugSaveFileReader{}
}

// Open implements the SaveFileReader interface
func (f *DebugSaveFileReader) Open(p string) error {
	f.savePath = filepath.Clean(p)
	info, err := os.Stat(f.savePath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", f.savePath)
	}
	return nil
}

// GetReader implements the SaveFileReader interface
func (f *DebugSaveFileReader) GetReader(segment string) (io.ReadCloser, error) {
	return os.Open(filepath.Clean(path.Join(f.savePath, segment)))
}

// Close implements the SaveFileReader interface
func (f *DebugSaveFileReader) Close() error {
	return nil
}

// DebugSaveFileWriter writes data sets to individual files in a directory.
type DebugSaveFileWriter struct {
	// Path to the directory for the save files
	savePath string
}

// NewDebugSaveFileWriter returns a new DebugSaveFileWriter object.
func NewDebugSaveFileWriter() *DebugSaveFileWriter {
	return &DebugSaveFileWriter{}
}

// Open implements the SaveFileWriter interface.
func (f *DebugSaveFileWriter) Open(p string) error {
	f.savePath = filepath.Clean(p)
	return os.MkdirAll(f.savePath, 0777)
}

// GetWriter implements the SaveFileWriter interface
func (f *DebugSaveFileWriter) GetWriter(segment string) (io.WriteCloser, error) {
	return os.Create(filepath.Clean(path.Join(f.savePath, segment)))
}

// Close implements the SaveFileWriter interface
func (f *DebugSaveFileWriter) Close() error {
	return nil
}
