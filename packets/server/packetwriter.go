package server

import (
	"encoding/binary"
)

// PacketWriter is a thin wrapper around a byte slice with methods specific to
// the UO wire protocol useful for encoding. Buf should be initialized with an
// empty slice.
type PacketWriter struct {
	Buf []byte
}

// PutByte writes an unsigned 8-bit byte
func (p *PacketWriter) PutByte(v byte) {
	p.Buf = append(p.Buf, v)
}

// PutUInt16 writes an unsigned 16-bit integer
func (p *PacketWriter) PutUInt16(v uint16) {
	l := len(p.Buf)
	binary.BigEndian.PutUint16(p.Buf[l:l+2], v)
	p.Buf = p.Buf[:l+2]
}

// PutUInt32 writes an unsigned 32-bit integer
func (p *PacketWriter) PutUInt32(v uint32) {
	l := len(p.Buf)
	binary.BigEndian.PutUint32(p.Buf[l:l+4], v)
	p.Buf = p.Buf[:l+4]
}

// PutASCII writes a fixed-length, zero-padded string
func (p *PacketWriter) PutASCII(v string, length int) {
	var i int
	for i, cp := range v {
		if i >= length {
			break
		}
		p.Buf = append(p.Buf, byte(cp&0x7f))
	}
	for ; i < length; i++ {
		p.Buf = append(p.Buf, 0)
	}
}

// PutBytes writes a byte slice
func (p *PacketWriter) PutBytes(b []byte) {
	p.Buf = append(p.Buf, b...)
}
