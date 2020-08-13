package uo

import "io"

const netCompressBufferLength = 128 * 1024

// ServerPacketWriter writes server (or client) packets to an io.Writer.
type ServerPacketWriter struct {
	w        io.Writer
	compress bool
	cbuf     []byte
}

// NewServerPacketWriter constructs a new ServerPacketWriter writing to w.
func NewServerPacketWriter(w io.Writer) *ServerPacketWriter {
	return &ServerPacketWriter{
		w: w,
	}
}

// SetCompression turns output compression on or off.
func (w *ServerPacketWriter) SetCompression(compress bool) {
	w.compress = compress
	if compress {
		if w.cbuf == nil {
			w.cbuf = make([]byte, netCompressBufferLength)
		}
	} else {
		w.cbuf = nil
	}
}

// WritePacket attempts to write packet p.
func (w *ServerPacketWriter) WritePacket(p ServerPacket) error {
	var buf []byte

	if w.compress {
		buf = HuffmanEncodePacket(p.Bytes(), w.cbuf[:0])
	} else {
		buf = p.Bytes()
	}
	_, err := w.w.Write(buf)
	return err
}
