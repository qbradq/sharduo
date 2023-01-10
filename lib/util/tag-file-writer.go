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
	ListFileWriter
}

// NewTagFileWriter returns a new TagFileWriter object ready to use for output.
func NewTagFileWriter(w io.WriteCloser) *TagFileWriter {
	return &TagFileWriter{
		ListFileWriter: *NewListFileWriter(w),
	}
}

// WriteObject writes a Serializeable to the given io.Writer.
func (f *TagFileWriter) WriteObject(s Serializeable) {
	f.WriteLine(fmt.Sprintf("[%s]", s.TypeName()))
	s.Serialize(f)
	f.WriteBlankLine()
}

// WriteNumber writes a number to the io.Writer in base 10.
func (f *TagFileWriter) WriteNumber(name string, value int) {
	f.WriteLine(fmt.Sprintf("%s=%d", name, value))
}

// WriteULong writes a 64-bit unsigned number to the io.Writer in base 10.
func (f *TagFileWriter) WriteULong(name string, value uint64) {
	f.WriteLine(fmt.Sprintf("%s=%d", name, value))
}

// WriteFloat writes a float to the io.Writer in base 10.
func (f *TagFileWriter) WriteFloat(name string, value float32) {
	f.WriteLine(fmt.Sprintf("%s=%f", name, value))
}

// WriteHex writes a number to the io.Writer in base 16 without leading zeros.
func (f *TagFileWriter) WriteHex(name string, value uint32) {
	f.WriteLine(fmt.Sprintf("%s=0x%X", name, value))
}

// WriteString writes a string to the io.Writer. Leading and trailing
// whitespace will be stripped. Empty strings are silently omitted.
func (f *TagFileWriter) WriteString(name, value string) {
	if value == "" {
		return
	}
	value = strings.TrimSpace(value)
	f.WriteLine(fmt.Sprintf("%s=%s", name, value))
}

// WriteBool writes a boolean value to the io.Writer. If value is false, no
// line will be emitted. Otherwise only the key name is emitted.
func (f *TagFileWriter) WriteBool(name string, value bool) {
	if value {
		f.WriteLine(name)
	}
}

// WriteLocation writes a uo.Location value to the io.Writer in tag file format.
func (f *TagFileWriter) WriteLocation(name string, l uo.Location) {
	f.WriteLine(fmt.Sprintf("%s=%d,%d,%d", name, l.X, l.Y, l.Z))
}

// WriteBounds writes a uo.Bounds value to the io.Writer in tag file format.
func (f *TagFileWriter) WriteBounds(name string, b uo.Bounds) {
	f.WriteLine(fmt.Sprintf("%s=%d,%d,%d,%d", name, b.X, b.Y, b.W, b.H))
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

// WriteObjectReferences writes a slice of serial values to the io.Writer. This
// uses line continuations in the tag file to avoid overrunning scanner buffers.
func (f *TagFileWriter) WriteObjectReferences(name string, ss []uo.Serial) {
	// Empty slices are omitted, the default value for an object reference list
	// is a nil slice.
	if len(ss) == 0 {
		return
	}

	// Build the property string
	b := strings.Builder{}
	b.WriteString(name)
	b.WriteRune('=')
	for idx, s := range ss {
		// Write the separator if needed
		if idx != 0 {
			// Write a line continuation every eight serial
			if idx%8 == 0 {
				b.Write([]byte(",\\\n"))
			} else {
				b.WriteRune(',')
			}
		}
		// Write the serial value
		b.WriteString(s.String())
	}
	f.WriteLine(b.String())
}
