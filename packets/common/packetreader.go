package common

// PacketReader is a thin wrapper around a byte slice with methods specific to
// the UO wire protocol useful for decoding
type PacketReader struct {
	Buf []byte
	ofs int
}

// Seek sets the absolute position of the data cursor
func (p *PacketReader) Seek(pos int) {
	p.ofs = pos
}

// GetByte gets the next byte in the packet buffer
func (p *PacketReader) GetByte() byte {
	p.ofs++
	return p.Buf[p.ofs-1]
}

// GetASCII gets a fixed-length ASCII string
func (p *PacketReader) GetASCII(length int) string {
	p.ofs += length
	return string(p.Buf[p.ofs-length : p.ofs])
}
