package uod

import (
	"errors"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// These functions respond to inbound client packets within the network read
// goroutine directly to offload that work from the world goroutine. Note that
// these functions CANNOT interact with the world memory model.
func init() {
	embeddedHandlers.Add(0x09, ignorePacket) // Single click
	embeddedHandlers.Add(0x73, handleClientPing)
	embeddedHandlers.Add(0xad, handleClientSpeech)
	embeddedHandlers.Add(0xbd, handleClientVersion)
	embeddedHandlers.Add(0xc8, handleClientViewRange)
}

// PacketContext represents the context in which a packet may enter the server
type PacketContext struct {
	// The net state associated with the packet, if any. System packets tend
	// not to have net states attached.
	NetState *NetState
	// The client packet
	Packet clientpacket.Packet
}

var embeddedHandlers = util.NewRegistry[uo.Serial, func(*PacketContext)]("client-packets")

func ignorePacket(c *PacketContext) {
	// Do nothing
}

func handleClientPing(c *PacketContext) {
	p := c.Packet.(*clientpacket.Ping)
	c.NetState.Send(&serverpacket.Ping{
		Key: p.Key,
	})
}

func handleClientSpeech(c *PacketContext) {
	p := c.Packet.(*clientpacket.Speech)
	if len(p.Text) == 0 {
		return
	}
	if len(p.Text) > 1 {
		if p.Text[0] == '[' {
			cmd := ParseCommand(p.Text[1:])
			if cmd != nil {
				cmd.Compile()
				world.SendRequest(&SpeechCommandRequest{
					BaseWorldRequest: BaseWorldRequest{
						NetState: c.NetState,
					},
					Command: cmd,
				})
				return
			}
		}
	}
	if c.NetState != nil && c.NetState.m != nil {
		GlobalChat(c.NetState.m.DisplayName(), p.Text)
	}
}

func handleClientVersion(c *PacketContext) {
	p := c.Packet.(*clientpacket.Version)
	if p.String != "7.0.15.1" {
		c.NetState.Error("version check", errors.New("bad client version"))
	}
}

func handleClientViewRange(c *PacketContext) {
	p := c.Packet.(*clientpacket.ClientViewRange)
	c.NetState.viewRange = uo.BoundViewRange(p.Range)
	c.NetState.Send(&serverpacket.ClientViewRange{
		Range: byte(c.NetState.ViewRange()),
	})
	// TODO Update visible objects for the client
}
