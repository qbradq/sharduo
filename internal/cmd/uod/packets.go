package uod

import (
	"errors"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	clientPacketFactory.add(0x02, handleWalkRequest)
	clientPacketFactory.add(0x06, ignorePacket)
	clientPacketFactory.add(0x09, ignorePacket)
	clientPacketFactory.add(0x73, handleClientPing)
	clientPacketFactory.add(0xad, handleClientSpeech)
	clientPacketFactory.add(0xbd, handleClientVersion)
	clientPacketFactory.add(0xc8, handleClientViewRange)
}

// PacketContext represents the context in which a packet may enter the server
type PacketContext struct {
	// The net state associated with the packet, if any. System packets tend
	// not to have net states attached.
	NetState *NetState
	// The client packet
	Packet clientpacket.Packet
}

var clientPacketFactory = util.NewFactory[int, *PacketContext]("client-packets")

func ignorePacket(c *PacketContext) {
	// Do nothing
}

func handleClientPing(c *PacketContext) {
	p := c.Packet.(*clientpacket.Ping)
	n.Send(&serverpacket.Ping{
		Key: p.Key,
	})
}

func handleClientSpeech(c *PacketContext) {
	p := c.Packet.(*clientpacket.Speech)
	if len(p.Text) == 0 {
		return
	}
	if p.Text[0] == '[' {

	}
	if n.m != nil {
		GlobalChat(n.m.GetDisplayName(), p.Text)
	}
}

func handleClientVersion(c *PacketContext) {
	p := c.Packet.(*clientpacket.Version)
	if p.String != "7.0.15.1" {
		n.Error("version check", errors.New("bad client version"))
	}
}

func handleClientViewRange(c *PacketContext) {
	p := c.Packet.(*clientpacket.ClientViewRange)
	n.Send(&serverpacket.ClientViewRange{
		Range: byte(p.Range),
	})
}

func handleWalkRequest(c *PacketContext) {
	p := c.Packet.(*clientpacket.WalkRequest)
	n.Send(&serverpacket.MoveAcknowledge{
		Sequence:  p.Sequence,
		Notoriety: uo.NotorietyInnocent,
	})
}
