package uod

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/clientpacket"
)

// PacketHandler is the signature of a function that responds to a client
// packet dispatched from PacketHandlerTable.
type PacketHandler func(n *NetState, p clientpacket.Packet)

// PacketHandlerFactory manages a collection of packet handlers.
type PacketHandlerFactory struct {
	// Map of all packet handlers
	handlers map[int]PacketHandler
	// Name of this factory for logging purposes
	name string
}

func newPacketHandlerFactory(name string) *PacketHandlerFactory {
	return &PacketHandlerFactory{
		handlers: make(map[int]PacketHandler),
		name:     name,
	}
}

func (f *PacketHandlerFactory) add(id int, handler PacketHandler) {
	if _, duplicate := f.handlers[id]; duplicate {
		panic(fmt.Sprintf("duplicate %s packet handler %d", f.name, id))
	}
	f.handlers[id] = handler
}

func (f *PacketHandlerFactory) get(id int) PacketHandler {
	return f.handlers[id]
}
