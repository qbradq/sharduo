package clientpacket

import (
	"log"
)

// packetFactory creates client packets from slices of bytes
type packetFactory struct {
	// Packet constructors
	ctors []func([]byte) Packet
}

// Add adds a constructor to the factory.
func (f *packetFactory) Add(id byte, ctor func([]byte) Packet) {
	if f.ctors == nil {
		f.ctors = make([]func([]byte) Packet, 0xFF)
	}
	if f.ctors[id] != nil {
		log.Fatalf("duplicate packet ctor for packet %02X", id)
	}
	f.ctors[id] = ctor
}

// Ignore ignores the given packet ID
func (f *packetFactory) Ignore(id byte) {
	f.Add(id, func(in []byte) Packet {
		p := &IgnoredPacket{
			basePacket: basePacket{id: id},
		}
		return p
	})
}

// New creates a new client packet.
func (f *packetFactory) New(id byte, in []byte) Packet {
	if id > 0xFF {
		log.Printf("error: packet id %08X out of range", id)
		return nil
	}
	ctor := f.ctors[id]
	if ctor == nil {
		up := NewUnsupportedPacket("client", in)
		up.basePacket.id = id
		return up
	}
	p := ctor(in)
	if p == nil {
		up := NewUnsupportedPacket("client-packets", in)
		up.basePacket.id = id
		return up
	}
	return p
}
