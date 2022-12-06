package util

import (
	"fmt"
	"io"
	"strings"

	"github.com/qbradq/sharduo/lib/uo"
)

// TagFileWriter reads and writes tag files
type TagFileWriter struct {
	w    io.Writer
	errs []error
}

// NewTagFileWriter returns a new TagFileWriter object ready to use for output.
func NewTagFileWriter(w io.Writer) *TagFileWriter {
	return &TagFileWriter{
		w: w,
	}
}

// handleError handles an error object.
func (f *TagFileWriter) handleError(err error) {
	f.errs = append(f.errs, err)
}

// Errors returns the slice of all errors encountered by this writer. A nil
// slice indicates no errors.
func (f *TagFileWriter) Errors() []error {
	return f.errs
}

// WriteCommentLine writes a single comment line to the tag file.
func (f *TagFileWriter) WriteCommentLine(comment string) {
	if _, err := f.w.Write([]byte(fmt.Sprintf("; %s\n", comment))); err != nil {
		f.handleError(err)
	}
}

// WriteObject writes a Serializeable to the given io.Writer.
func (f *TagFileWriter) WriteObject(s Serializeable) {
	if _, err := f.w.Write([]byte(fmt.Sprintf("\n[%s]\n", s.GetTypeName()))); err != nil {
		f.handleError(err)
	}
	s.Serialize(f)
}

// WriteNumber writes a number to the io.Writer in base 10.
func (f *TagFileWriter) WriteNumber(name string, value int) {
	if _, err := f.w.Write([]byte(fmt.Sprintf("%s=%d\n", name, value))); err != nil {
		f.handleError(err)
	}
}

// WriteHex writes a number to the io.Writer in base 16 without leading zeros.
func (f *TagFileWriter) WriteHex(name string, value int) {
	if _, err := f.w.Write([]byte(fmt.Sprintf("%s=0x%X\n", name, value))); err != nil {
		f.handleError(err)
	}
}

// WriteString writes a string to the io.Writer. Leading and trailing
// whitespace will be stripped. Empty strings are silently omitted.
func (f *TagFileWriter) WriteString(name, value string) {
	if value == "" {
		return
	}
	value = strings.TrimSpace(value)
	if _, err := f.w.Write([]byte(fmt.Sprintf("%s=%s\n", name, value))); err != nil {
		f.handleError(err)
	}
}

// WriteBool writes a boolean value to the io.Writer. If value is false, no
// line will be emitted. Otherwise only the key name is emitted.
func (f *TagFileWriter) WriteBool(name string, value bool) {
	if value {
		if _, err := f.w.Write([]byte(fmt.Sprintf("%s\n", name))); err != nil {
			f.handleError(err)
		}
	}
}

// WriteSerialSlice writes a slice of serial values to the io.Writer.
func (f *TagFileWriter) WriteSerialSlice(name string, ss []uo.Serial) {
	b := strings.Builder{}
	if _, err := b.WriteString(name); err != nil {
		panic(err)
	}
	if _, err := b.WriteRune('='); err != nil {
		panic(err)
	}
	for idx, s := range ss {
		if idx == 0 {
			b.WriteString(s.String())
		} else {
			b.WriteString(", " + s.String())
		}
	}
}
