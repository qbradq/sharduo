package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// Global buffers
var objectsINIBuf = bytes.NewBuffer(make([]byte, 1024*1024*512))
var mapINIBuf = bytes.NewBuffer(make([]byte, 1024*1024*16))

// NewSaveFileReaderByPath creates a new SaveFileReader implementation based on
// the characteristics of the path.
//
// Characteristics:
// 1. If the path has no file extension it is assumed to be a flat file save.
// 2. If the path has the file extension ".zip" it is assumed to be compressed.
// 3. An error is returned in any other case.
//   - If the path does not exist this will return an error.
func NewSaveFileReaderByPath(p string) (SaveFileReader, error) {
	if _, err := os.Stat(p); err != nil {
		return nil, err
	}
	ext := path.Ext(p)
	switch ext {
	case "":
		return NewSaveFileReader("Flat")
	case ".zip":
		return NewSaveFileReader("Compressed")
	default:
		return nil, fmt.Errorf("unsupported save file type %s", ext)
	}
}

// NewSaveFileReader creates a new SaveFileReader implementation by name.
func NewSaveFileReader(which string) (SaveFileReader, error) {
	switch which {
	case "Flat":
		return NewFlatSaveFileReader(), nil
	case "Compressed":
		return NewCompressedSaveFileReader(), nil
	default:
		return nil, fmt.Errorf("unknown SaveFileReader type %s", which)
	}
}

// NewSaveFileWriter creates a new SaveFileWriter implementation by name.
func NewSaveFileWriter(which string) (SaveFileWriter, error) {
	switch which {
	case "Flat":
		return NewFlatSaveFileWriter(), nil
	case "Compressed":
		return NewCompressedSaveFileWriter(), nil
	default:
		return nil, fmt.Errorf("unknown SaveFileWriter type %s", which)
	}
}

// writeCloserWrapper is a wrapper struct for an io.Writer with a do-nothing
// Close() function so it implements the io.WriteCloser interface.
type writeCloserWrapper struct {
	w io.Writer
}

// Write implements the io.Writer interface.
func (w writeCloserWrapper) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

// Close implements the io.Closer interface.
func (w writeCloserWrapper) Close() error { return nil }

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
func (f *CompressedSaveFileWriter) GetWriter(segment string) (io.WriteCloser, error) {
	o, err := f.w.Create(segment)
	if err != nil {
		return nil, err
	}
	return writeCloserWrapper{w: o}, nil
}

// Close implements the SaveFileWriter interface.
func (f *CompressedSaveFileWriter) Close() error {
	return f.w.Close()
}

// FlatSaveFileReader reads data sets from individual files from a directory.
type FlatSaveFileReader struct {
	// Path to the directory for the save files
	savePath string
}

// NewFlatSaveFileReader creates a new DebugSaveFileReader and returns it.
func NewFlatSaveFileReader() *FlatSaveFileReader {
	return &FlatSaveFileReader{}
}

// Open implements the SaveFileReader interface
func (f *FlatSaveFileReader) Open(p string) error {
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
func (f *FlatSaveFileReader) GetReader(segment string) (io.ReadCloser, error) {
	return os.Open(filepath.Clean(path.Join(f.savePath, segment)))
}

// Close implements the SaveFileReader interface
func (f *FlatSaveFileReader) Close() error {
	return nil
}

// FlatSaveFileWriter writes data sets to individual files in a directory.
type FlatSaveFileWriter struct {
	// Path to the directory for the save files
	savePath string
	// Map of all output buffers
	bufs map[string]*bytes.Buffer
}

// NewFlatSaveFileWriter returns a new DebugSaveFileWriter object.
func NewFlatSaveFileWriter() *FlatSaveFileWriter {
	return &FlatSaveFileWriter{
		bufs: make(map[string]*bytes.Buffer),
	}
}

// Open implements the SaveFileWriter interface.
func (f *FlatSaveFileWriter) Open(p string) error {
	f.savePath = filepath.Clean(p)
	return os.MkdirAll(f.savePath, 0777)
}

// GetWriter implements the SaveFileWriter interface
func (f *FlatSaveFileWriter) GetWriter(segment string) (io.WriteCloser, error) {
	if _, duplicate := f.bufs[segment]; duplicate {
		return nil, fmt.Errorf("duplicate segment %s", segment)
	}
	var buf *bytes.Buffer
	// Hacky for now
	switch segment {
	case "objects.ini":
		buf = objectsINIBuf
	case "map.ini":
		buf = mapINIBuf
	default:
		buf = bytes.NewBuffer(make([]byte, 1024*64))
	}
	buf.Reset()
	f.bufs[segment] = buf
	return &writeCloserWrapper{
		w: buf,
	}, nil
}

// Close implements the SaveFileWriter interface
func (f *FlatSaveFileWriter) Close() error {
	for p, buf := range f.bufs {
		if err := os.WriteFile(path.Join(f.savePath, p), buf.Bytes(), 0777); err != nil {
			return err
		}
	}
	return nil
}
