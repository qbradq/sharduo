package marshal

import (
	"bytes"
	"encoding/binary"
)

// Writer is a wrapper around bytes.Buffer that has some extra helper functions
// for writing multi-byte values in little-endian format. The zero value of
// Writer is valid and ready for operation.
type Writer struct {
	bytes.Buffer
	// Internal temp buffer
	buf []byte
}

// WriteShort writes a 16-bit signed integer to the buffer.
func (w *Writer) WriteShort(v int16) {
	binary.LittleEndian.PutUint16(w.buf, uint16(v))
	w.Write(w.buf[:2])
}

// WriteInt writes a 32-bit signed integer to the buffer.
func (w *Writer) WriteInt(v int32) {
	binary.LittleEndian.PutUint32(w.buf, uint32(v))
	w.Write(w.buf[:4])
}

// WriteLong writes a 64-bit signed integer to the buffer.
func (w *Writer) WriteLong(v int64) {
	binary.LittleEndian.PutUint32(w.buf, uint32(v))
	w.Write(w.buf[:8])
}

// WriteString writes a null-terminates string to the buffer.
func (w *Writer) WriteString(s string) {
	w.Buffer.WriteString(s)
	w.WriteByte(0)
}
