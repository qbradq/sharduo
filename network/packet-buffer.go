package network

// PacketBuffer is a thin wrapper around a byte slice with methods specific to
// the UO wire protocol
type PacketBuffer struct {
	Buf []byte
	ofs int
}

// Seek sets the absolute position of the data cursor
func (p *PacketBuffer) Seek(pos int) {
	p.ofs = pos
}

// GetByte gets the next byte in the packet buffer
func (p *PacketBuffer) GetByte() byte {
	p.ofs++
	return p.Buf[p.ofs-1]
}

// GetASCII gets a fixed-length ASCII string
func (p *PacketBuffer) GetASCII(length int) string {
	p.ofs += length
	return string(p.Buf[p.ofs-length : p.ofs])
}
