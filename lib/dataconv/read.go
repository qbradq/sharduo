package dataconv

import (
	"encoding/binary"
	"unicode/utf16"
)

func NullString(buf []byte) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

func UTF16String(b []byte) string {
	var utf [256]uint16
	ib := 0
	iu := 0
	for {
		if ib+1 >= len(b) || iu >= len(utf) {
			break
		}
		if b[ib+1] == 0 && b[ib] == 0 {
			break
		}
		utf[iu] = binary.BigEndian.Uint16(b[ib:])
		ib += 2
		iu++
	}
	return string(utf16.Decode(utf[:iu]))
}

func GetUint16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

func GetUint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}
