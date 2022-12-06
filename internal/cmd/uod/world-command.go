package uod

import (
	"github.com/qbradq/sharduo/lib/clientpacket"
)

// CommandType are the type codes for commands
type CommandType int

// All valid values for CommandType
const (
	CommandTypeSpeechCommand int = 0x0100
)

func init() {
	worldCommandFactory.add(CommandTypeSpeechCommand, handleSpeechCommand)
}

// Factory for world commands
var worldCommandFactory = newPacketHandlerFactory("world-commands")

// WorldCommand is used to send client and system packets to the world's
// goroutine.
type WorldCommand struct {
	// The net state associated with the command, if any. System commands tend
	// not to have associated net states.
	NetState *NetState
	// The client or system packet associated with this command.
	Packet clientpacket.Packet
}

func handleSpeechCommand(n *NetState, p clientpacket.Packet) {
	sp := p.(*clientpacket.Speech)
	c := ParseCommand(sp.Text)
}
