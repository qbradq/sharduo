package serverpacket

import (
	"bytes"
	"io"

	"github.com/qbradq/sharduo/lib/uo"
)

// CompressedWriter wraps the process of writing compressed server packets.
type CompressedWriter struct {
	cpBuf *bytes.Buffer
}

// NewCompressedWriter returns a CompressedWriter ready for use.
func NewCompressedWriter() *CompressedWriter {
	return &CompressedWriter{
		cpBuf: bytes.NewBuffer(nil),
	}
}

// Write writes p to w with compression and returns any error from w.Write().
func (c *CompressedWriter) Write(p Packet, w io.Writer) error {
	c.cpBuf.Reset()
	p.Write(c.cpBuf)
	return uo.HuffmanEncodePacket(c.cpBuf, w)
}
