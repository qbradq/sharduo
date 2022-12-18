package util

import (
	"io"
)

// ListFileWriter writes data to a writer in list file format
type ListFileWriter struct {
	w io.WriteCloser
}

// NewListFileWriter creates a new ListFileWriter for the writer. Be sure to
// call ListFileWriter.Close().
func NewListFileWriter(w io.WriteCloser) *ListFileWriter {
	return &ListFileWriter{
		w: w,
	}
}

// WriteBlankLine writes a blank line
func (f *ListFileWriter) WriteBlankLine() {
	f.w.Write([]byte("\n"))
}

// WriteComment writes a comment line
func (f *ListFileWriter) WriteComment(line string) {
	f.w.Write([]byte("; " + line + "\n"))
}

// WriteSegmentHeader writes a segment header for the named segment
func (f *ListFileWriter) WriteSegmentHeader(name string) {
	f.w.Write([]byte("[" + name + "]\n"))
}

// WriteLine writes a single line to the list file
func (f *ListFileWriter) WriteLine(line string) {
	f.w.Write([]byte(line + "\n"))
}

// Close closes the list file
func (f *ListFileWriter) Close() error {
	return f.w.Close()
}
