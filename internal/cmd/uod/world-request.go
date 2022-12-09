package uod

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/clientpacket"
)

// WorldRequest is used to send client and system packets to the world's
// goroutine.
type WorldRequest interface {
	// Returns the NetState associated with this request, if any
	GetNetState() *NetState
	// Returns the Packet associated with this request, if any
	GetPacket() clientpacket.Packet
	// Execute executes the request
	Execute() error
}

// BaseWorldRequest provides the base implementation of WorldRequest except for
// Execute and GetID to force includers to provide their own.
type BaseWorldRequest struct {
	// The net state associated with the command, if any. System commands tend
	// not to have associated net states.
	NetState *NetState
	// The client or system packet associated with this command.
	Packet clientpacket.Packet
}

// GetNetState implements the WorldRequest interface
func (r *BaseWorldRequest) GetNetState() *NetState {
	return r.NetState
}

// GetPacket implements the WorldRequest interface
func (r *BaseWorldRequest) GetPacket() clientpacket.Packet {
	return r.Packet
}

// ClientPacketRequest is sent by the NetState for packets that should be
// addressed directly in the world goroutine.
type ClientPacketRequest struct {
	BaseWorldRequest
}

// Execute implements the WorldRequest interface
func (r *ClientPacketRequest) Execute() error {
	handler, found := worldHandlers.Get(r.GetPacket().GetSerial())
	if !found || handler == nil {
		return fmt.Errorf("unhandled packet 0x%08X", int(r.GetPacket().GetSerial()))
	}
	handler(r.GetNetState(), r.GetPacket())
	return nil
}

// SpeechCommandRequest is sent by a player using the speech packet with the
// message starting with the '[' character
type SpeechCommandRequest struct {
	BaseWorldRequest
	Command Command
}

// Executes implements the WorldRequest interface
func (r *SpeechCommandRequest) Execute() error {
	return r.Command.Execute(r.NetState)
}
