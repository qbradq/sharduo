package serverpacket

import (
	"bytes"
	"io"

	"github.com/qbradq/sharduo/lib/uo"
)

// CompressedWriter wraps the process of writting compressed server packets.
type CompressedWriter struct {
	cpbuf *bytes.Buffer
}

// NewCompressedWriter returns a CompressedWriter ready for use.
func NewCompressedWriter() *CompressedWriter {
	return &CompressedWriter{
		cpbuf: bytes.NewBuffer(nil),
	}
}

// Write writes p to w with compression and returns any error from w.Write().
func (c *CompressedWriter) Write(p Packet, w io.Writer) error {
	c.cpbuf.Reset()
	p.Write(c.cpbuf)
	return uo.HuffmanEncodePacket(c.cpbuf, w)
}
