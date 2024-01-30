package util

import (
	"encoding/binary"
	"io"
	"math"
	"time"
	"unicode/utf16"

	"github.com/qbradq/sharduo/lib/uo"
)

// utf16Buf is an internal buffer for various routines.
var utf16Buf [1024 * 16]uint16

// dcBuf is an internal buffer for various routines.
var dcBuf [1024 * 16]byte

// ParseNullString returns the next null-terminated string from the byte slice.
func ParseNullString(buf []byte) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

// ParseUTF16String returns the next UTF16-encoded string from the byte slice.
func ParseUTF16String(b []byte) string {
	ib := 0
	iu := 0
	for {
		if ib+1 >= len(b) || iu >= len(utf16Buf) {
			break
		}
		if b[ib+1] == 0 && b[ib] == 0 {
			break
		}
		utf16Buf[iu] = binary.BigEndian.Uint16(b[ib:])
		ib += 2
		iu++
	}
	return string(utf16.Decode(utf16Buf[:iu]))
}

// ParseUInt16 returns the next 16-bit unsigned integer in the byte slice.
func ParseUInt16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

// ParseUInt32 returns the next 32-bit unsigned integer in the byte slice.
func ParseUInt32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

// PutBool writes a boolean value
func PutBool(w io.Writer, v bool) {
	var b [1]byte
	if v {
		b[0] = 1
	} else {
		b[0] = 0
	}
	w.Write(b[:])
}

// PutString writes a null-terminated string
func PutString(w io.Writer, s string) {
	dcBuf[0] = 0
	w.Write([]byte(s))
	w.Write(dcBuf[:1])
}

// Writes a fixed-length string
func PutStringN(w io.Writer, s string, n int) {
	copy(dcBuf[:n], s)
	w.Write(dcBuf[:n])
}

// Writes a fixed-length string that always ends with a null
func PutStringNWithNull(w io.Writer, s string, n int) {
	copy(dcBuf[:n], s)
	dcBuf[n-1] = 0
	w.Write(dcBuf[:n])
}

// PutUTF16String writes a null-terminated UTF16 string in Big Endian format
func PutUTF16String(w io.Writer, s string) {
	utf := utf16.Encode([]rune(s))
	ofs := 0
	for _, r := range utf {
		binary.BigEndian.PutUint16(dcBuf[ofs:ofs+2], r)
		ofs += 2
	}
	dcBuf[ofs] = 0
	dcBuf[ofs+1] = 0
	w.Write(dcBuf[:ofs+2])
}

// Writes a UTF16 string in Big Endian format with no terminator
func PutUTF16StringN(w io.Writer, s string, n int) {
	utf := utf16.Encode([]rune(s))
	for idx := 0; idx < n; idx++ {
		r := uint16(0)
		if idx < len(utf) {
			r = utf[idx]
		}
		binary.BigEndian.PutUint16(dcBuf[idx*2:idx*2+2], r)
	}
	w.Write(dcBuf[:n*2])
}

// Writes a null-terminated UTF16 string in Little Endian format
func PutUTF16LEString(w io.Writer, s string) {
	utf := utf16.Encode([]rune(s))
	ofs := 0
	for _, r := range utf {
		binary.LittleEndian.PutUint16(dcBuf[ofs*2:ofs*2+2], r)
	}
	dcBuf[ofs] = 0
	dcBuf[ofs+1] = 0
	w.Write(dcBuf[:ofs+2])
}

// Writes a UTF16 string in Little Endian format with no terminator
func PutUTF16LEStringN(w io.Writer, s string, n int) {
	utf := utf16.Encode([]rune(s))
	for idx := 0; idx < n; idx++ {
		r := uint16(0)
		if idx < len(utf) {
			r = utf[idx]
		}
		binary.LittleEndian.PutUint16(dcBuf[idx*2:idx*2+2], r)
	}
	w.Write(dcBuf[:n*2])
}

// Pad writes zero-padding
func Pad(w io.Writer, l int) {
	for i := 0; i < l; i++ {
		dcBuf[i] = 0
	}
	w.Write(dcBuf[:l])
}

// Fill fills with a given byte
func Fill(w io.Writer, v byte, l int) {
	for i := 0; i < l; i++ {
		dcBuf[i] = v
	}
	w.Write(dcBuf[:l])
}

// PutByte writes a single byte
func PutByte(w io.Writer, v byte) {
	dcBuf[0] = v
	w.Write(dcBuf[:1])
}

// PutUInt16 writes a 16-bit numeric value
func PutUInt16(w io.Writer, v uint16) {
	binary.BigEndian.PutUint16(dcBuf[:2], v)
	w.Write(dcBuf[:2])
}

// PutUInt32 writes a 32-bit numeric value
func PutUInt32(w io.Writer, v uint32) {
	binary.BigEndian.PutUint32(dcBuf[:4], v)
	w.Write(dcBuf[:4])
}

// PutUInt64 writes a 64-bit numeric value
func PutUInt64(w io.Writer, v uint64) {
	binary.BigEndian.PutUint64(dcBuf[:8], v)
	w.Write(dcBuf[:8])
}

// PutFloat writes a 64-bit floating-point value.
func PutFloat(w io.Writer, v float64) {
	PutUInt64(w, math.Float64bits(v))
}

// PutPoint writes a Point value to the writer.
func PutPoint(w io.Writer, p uo.Point) {
	PutUInt16(w, uint16(p.X))
	PutUInt16(w, uint16(p.Y))
	PutByte(w, byte(int8(p.Z)))
}

// PutBounds writes a Rect value to the writer.
func PutBounds(w io.Writer, r uo.Bounds) {
	PutUInt16(w, uint16(r.X))
	PutUInt16(w, uint16(r.Y))
	PutByte(w, byte(int8(r.Z)))
	PutUInt16(w, uint16(r.W))
	PutUInt16(w, uint16(r.H))
	PutUInt16(w, uint16(r.D))
}

// PutTime writes a time.Time value to the writer.
func PutTime(w io.Writer, t time.Time) {
	PutUInt64(w, uint64(t.UnixMilli()))
}

// PutBytes puts a slice of bytes to the writer.
func PutBytes(w io.Writer, d []byte) {
	PutUInt32(w, uint32(len(d)))
	w.Write(d)
}

// GetBool returns the next boolean value in the data buffer.
func GetBool(r io.Reader) bool {
	return GetByte(r) != 0
}

// GetString returns the next null-terminated string in the data buffer.
func GetString(r io.Reader) string {
	var buf = dcBuf[:1]
	var ret []byte
	for {
		r.Read(buf)
		if buf[0] == 0 {
			return string(ret)
		}
		ret = append(ret, buf[0])
	}
}

// GetByte is a convenience function that returns the next byte in the buffer.
func GetByte(r io.Reader) byte {
	var buf = dcBuf[:1]
	r.Read(buf)
	return buf[0]
}

// GetUInt16 returns the next unsigned 16-bit integer in the data buffer.
func GetUInt16(r io.Reader) uint16 {
	var buf = dcBuf[:2]
	r.Read(buf)
	return binary.BigEndian.Uint16(buf)
}

// GetUInt32 returns the next unsigned 32-bit integer in the data buffer.
func GetUInt32(r io.Reader) uint32 {
	var buf = dcBuf[:4]
	r.Read(buf)
	return binary.BigEndian.Uint32(buf)
}

// GetUInt64 returns the next unsigned 64-bit integer in the data buffer.
func GetUInt64(r io.Reader) uint64 {
	var buf = dcBuf[:8]
	r.Read(buf)
	return binary.BigEndian.Uint64(buf)
}

// GetFloat returns the next 64-bit floating-point value.
func GetFloat(r io.Reader) float64 {
	return math.Float64frombits(GetUInt64(r))
}

// GetPoint returns the next Point value in the data buffer.
func GetPoint(r io.Reader) uo.Point {
	return uo.Point{
		X: int(GetUInt16(r)),
		Y: int(GetUInt16(r)),
		Z: int(int8(GetByte(r))),
	}
}

// GetBounds returns the next Bounds value in the data buffer.
func GetRect(r io.Reader) uo.Bounds {
	return uo.Bounds{
		X: int(GetUInt16(r)),
		Y: int(GetUInt16(r)),
		Z: int(int8(GetByte(r))),
		W: int(GetUInt16(r)),
		H: int(GetUInt16(r)),
		D: int(GetUInt16(r)),
	}
}

// GetTime returns the next time.Time value in the data buffer.
func GetTime(r io.Reader) time.Time {
	return time.UnixMilli(int64(GetUInt64(r)))
}

// GetBytes returns the next byte slice in the data buffer.
func GetBytes(r io.Reader) []byte {
	n := GetUInt32(r)
	ret := make([]byte, n)
	r.Read(ret)
	return ret
}
