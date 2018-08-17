package uo

import (
	"encoding/binary"
	"io"
)

// A ClientPacketReader reads Ultima Online client packets from an io.Reader
type ClientPacketReader struct {
	reader     io.Reader
	readBuffer []byte
}

// NewClientPacketReader creates a new ClientPacketReader with the given
// io.Reader as the source.
func NewClientPacketReader(r io.Reader) *ClientPacketReader {
	return &ClientPacketReader{
		reader:     r,
		readBuffer: make([]byte, 1024*64, 1024*64),
	}
}

// ReadConnectionHeader reads the 4-byte connection header used by Ultima Online
// tcp streams
func (r *ClientPacketReader) ReadConnectionHeader() error {
	_, err := io.ReadFull(r.reader, r.readBuffer[0:4])
	return err
}

// ReadClientPacket reads the next client packet from the underlying reader
func (r *ClientPacketReader) ReadClientPacket() (ClientPacket, error) {
	_, err := io.ReadFull(r.reader, r.readBuffer[0:1])
	if err != nil {
		return nil, err
	}

	info := clientPacketInfos[r.readBuffer[0]]
	length := info.length

	// Packet body
	if length == 0 { // Bad packet
		return ClientPacketInvalid(r.readBuffer), nil
	} else if length == -1 { // Dynamic length
		_, err := io.ReadFull(r.reader, r.readBuffer[1:3])
		if err != nil {
			return nil, err
		}
		length := binary.BigEndian.Uint16(r.readBuffer[1:3])
		_, err = io.ReadFull(r.reader, r.readBuffer[3:length])
		if err != nil {
			return nil, err
		}
	} else { // Fixed length
		_, err = io.ReadFull(r.reader, r.readBuffer[1:length])
		if err != nil {
			return nil, err
		}
	}

	// Packet object creation
	if info.decoder == nil {
		return ClientPacketNotSupported(r.readBuffer), nil
	}
	return info.decoder(r.readBuffer), nil
}
