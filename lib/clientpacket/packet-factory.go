package clientpacket

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
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
		p := &IgnoredPacket{}
		p.SetSerial(uo.Serial(id))
		return p
	})
}

// New creates a new client packet.
func (f *packetFactory) New(id uo.Serial, in []byte) Packet {
	if p := f.New(id, in); p != nil {
		return p
	}
	up := NewUnsupportedPacket("client-packets", in)
	up.SetSerial(id)
	return up
}
