package marshal

import (
	"bytes"
	"encoding/binary"
)

// Reader is a wrapper around bytes.Buffer that has some extra helper functions
// for reading multi-byte values in little-endian format. The zero value of
// Reader is not valid, use NewReader to initialize a reader with data.
type Reader struct {
	bytes.Buffer
}

// NewReader creates a new reader object for the data given.
func NewReader(d []byte) *Reader {
	return &Reader{
		Buffer: *bytes.NewBuffer(d),
	}
}

// ReadShort reads a 16-bit signed integer from the buffer.
func (w *Writer) ReadShort() int16 {
	w.Read(w.buf[:2])
	return int16(binary.LittleEndian.Uint16(w.buf[:2]))
}

// ReadInt reads a 32-bit signed integer from the buffer.
func (w *Writer) ReadInt() int32 {
	w.Read(w.buf[:4])
	return int32(binary.LittleEndian.Uint32(w.buf[:4]))
}

// ReadLong reads a 64-bit signed integer from the buffer.
func (w *Writer) ReadLong() int64 {
	w.Read(w.buf)
	return int64(binary.LittleEndian.Uint64(w.buf))
}

// ReadString reads a null-terminated string from the buffer.
func (w *Writer) ReadString() string {
	s, _ := w.Buffer.ReadString(0)
	if len(s) < 1 {
		return s
	}
	return s[:len(s)-1]
}
