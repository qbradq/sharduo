package clientpacket

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrUnknownPacket is returned when an unknown packet is encountered.
var ErrUnknownPacket = errors.New("Unknown packet")

const maxInputBuffer = 64 * 1024

// Reader reads the client packets and returns the bytes of the packet.
type Reader struct {
	inbuf []byte
	r     io.Reader
	// Header is the connection header, or nil if it has not been read.
	Header []byte
}

// NewReader creates a new Reader for use.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:     r,
		inbuf: make([]byte, maxInputBuffer, maxInputBuffer),
	}
}

// ReadConnectionHeader reads the 8-byte connection header used by Ultima Online
// tcp streams
func (r *Reader) ReadConnectionHeader() error {
	// Connection header
	r.Header = make([]byte, 4, 4)
	_, err := io.ReadFull(r.r, r.Header)
	return err
}

// Read reads the bytes of the next client packet and returns a slice of those
// bytes or an error.
func (r *Reader) Read() ([]byte, error) {
	_, err := io.ReadFull(r.r, r.inbuf[0:1])
	if err != nil {
		// io.EOF here means no more data waiting
		return nil, err
	}

	info := InfoTable[r.inbuf[0]]
	length := info.Length

	// Packet body
	if length == 0 { // Bad packet
		return nil, ErrUnknownPacket
	} else if length == -1 { // Dynamic length
		_, err := io.ReadFull(r.r, r.inbuf[1:3])
		if err != nil {
			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
		length = int(binary.BigEndian.Uint16(r.inbuf[1:3]))
		_, err = io.ReadFull(r.r, r.inbuf[3:length])
		if err != nil {
			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
	} else { // Fixed length
		_, err = io.ReadFull(r.r, r.inbuf[1:length])
		if err != nil {
			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
	}
	return r.inbuf[:length], nil
}

// ReadPacket reads the bytes as in Read but returns a client packet object.
func (r *Reader) ReadPacket() (Packet, error) {
	data, err := r.Read()
	if err != nil {
		return nil, err
	}
	return New(data), nil
}
