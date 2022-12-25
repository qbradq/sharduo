package util

import (
	"fmt"
	"io"
	"strings"

	"github.com/qbradq/sharduo/lib/uo"
	"golang.org/x/exp/maps"
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

// WriteBlankLine writes a single blank line to the tag file.
func (f *TagFileWriter) WriteBlankLine() {
	if _, err := f.w.Write([]byte("\n")); err != nil {
		f.handleError(err)
	}
}

// WriteObject writes a Serializeable to the given io.Writer.
func (f *TagFileWriter) WriteObject(s Serializeable) {
	if _, err := f.w.Write([]byte(fmt.Sprintf("\n[%s]\n", s.TypeName()))); err != nil {
		f.handleError(err)
	}
	s.Serialize(f)
	f.WriteBlankLine()
}

// WriteNumber writes a number to the io.Writer in base 10.
func (f *TagFileWriter) WriteNumber(name string, value int) {
	if _, err := f.w.Write([]byte(fmt.Sprintf("%s=%d\n", name, value))); err != nil {
		f.handleError(err)
	}
}

// WriteHex writes a number to the io.Writer in base 16 without leading zeros.
func (f *TagFileWriter) WriteHex(name string, value uint32) {
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

// WriteLocation writes a uo.Location value to the io.Writer in tag file format.
func (f *TagFileWriter) WriteLocation(name string, l uo.Location) {
	f.w.Write([]byte(fmt.Sprintf("%s=%d,%d,%d\n", name, l.X, l.Y, l.Z)))
}

// WriteBounds writes a uo.Bounds value to the io.Writer in tag file format.
func (f *TagFileWriter) WriteBounds(name string, b uo.Bounds) {
	f.w.Write([]byte(fmt.Sprintf("%s=%d,%d,%d,%d\n", name, b.X, b.Y, b.W, b.H)))
}

// ValuesAsSerials returns the values of the input map as a slice of uo.Serial
// values.
func ValuesAsSerials[K comparable, T Serialer](in map[K]T) []uo.Serial {
	vs := maps.Values(in)
	ret := make([]uo.Serial, 0, len(vs))
	for _, v := range vs {
		ret = append(ret, v.Serial())
	}
	return ret
}

// ToSerials returns the input slice converted to a slice of uo.Serial values.
func ToSerials[T Serialer](in []T) []uo.Serial {
	ret := make([]uo.Serial, 0, len(in))
	for _, s := range in {
		ret = append(ret, s.Serial())
	}
	return ret
}

// WriteObjectReferences writes a slice of serial values to the io.Writer.
func (f *TagFileWriter) WriteObjectReferences(name string, ss []uo.Serial) {
	// Empty slices are omitted, the default value for an object reference list
	// is a nil slice.
	if len(ss) == 0 {
		return
	}

	// Build the property string
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
			b.Write([]byte(","))
			b.WriteString(s.String())
		}
	}
	b.Write([]byte("\n"))

	// Write the property string to the tag file
	f.w.Write([]byte(b.String()))
}
