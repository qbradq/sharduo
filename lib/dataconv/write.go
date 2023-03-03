package dataconv

import (
	"encoding/binary"
	"io"
	"unicode/utf16"
)

// Writes a boolean value
func PutBool(w io.Writer, v bool) {
	var b [1]byte
	if v {
		b[0] = 1
	} else {
		b[0] = 0
	}
	w.Write(b[:])
}

// Writes a null-terminated string
func PutString(w io.Writer, s string) {
	var b [1]byte
	w.Write([]byte(s))
	w.Write(b[:])
}

// Writes a fixed-length string
func PutStringN(w io.Writer, s string, n int) {
	var b = make([]byte, n)
	copy(b, s)
	w.Write(b)
}

// Writes zero-padding
func Pad(w io.Writer, l int) {
	var buf [1024]byte
	w.Write(buf[:l])
}

// Fills with a given byte
func Fill(w io.Writer, v byte, l int) {
	var b [1]byte
	b[0] = v
	for i := 0; i < l; i++ {
		w.Write(b[:])
	}
}

// Writes a single byte
func PutByte(w io.Writer, v byte) {
	var b [1]byte
	b[0] = v
	w.Write(b[:])
}

// Writes a 16-bit numeric value
func PutUint16(w io.Writer, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	w.Write(b[:])
}

// Writes a 32-bit numeric value
func PutUint32(w io.Writer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	w.Write(b[:])
}

// Writes a null-terminated UTF16 string in Big Endian format
func PutUTF16String(w io.Writer, s string) {
	var zeroBuf [2]byte
	var buf [2]byte
	utf := utf16.Encode([]rune(s))
	for _, r := range utf {
		binary.BigEndian.PutUint16(buf[:], r)
		w.Write(buf[:])
	}
	w.Write(zeroBuf[:])
}

// Writes a UTF16 string in Big Endian format with no terminator
func PutUTF16StringN(w io.Writer, s string, n int) {
	var buf [2]byte
	utf := utf16.Encode([]rune(s))
	for idx, r := range utf {
		if idx >= n {
			break
		}
		binary.BigEndian.PutUint16(buf[:], r)
		w.Write(buf[:])
	}
}
