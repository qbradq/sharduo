package uod

import (
	"errors"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	clientPacketFactory.Add(0x02, handleWalkRequest)
	clientPacketFactory.Add(0x06, ignorePacket)
	clientPacketFactory.Add(0x09, ignorePacket)
	clientPacketFactory.Add(0x73, handleClientPing)
	clientPacketFactory.Add(0xad, handleClientSpeech)
	clientPacketFactory.Add(0xbd, handleClientVersion)
	clientPacketFactory.Add(0xc8, handleClientViewRange)
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

func ignorePacket(c *PacketContext) any {
	// Do nothing
	return nil
}

func handleClientPing(c *PacketContext) any {
	p := c.Packet.(*clientpacket.Ping)
	c.NetState.Send(&serverpacket.Ping{
		Key: p.Key,
	})
	return nil
}

func handleClientSpeech(c *PacketContext) any {
	p := c.Packet.(*clientpacket.Speech)
	if len(p.Text) == 0 {
		return nil
	}
	if p.Text[0] == '[' {

	}
	if c.NetState != nil && c.NetState.m != nil {
		GlobalChat(c.NetState.m.GetDisplayName(), p.Text)
	}
	return nil
}

func handleClientVersion(c *PacketContext) any {
	p := c.Packet.(*clientpacket.Version)
	if p.String != "7.0.15.1" {
		c.NetState.Error("version check", errors.New("bad client version"))
	}
	return nil
}

func handleClientViewRange(c *PacketContext) any {
	p := c.Packet.(*clientpacket.ClientViewRange)
	c.NetState.Send(&serverpacket.ClientViewRange{
		Range: byte(p.Range),
	})
	return nil
}

func handleWalkRequest(c *PacketContext) any {
	p := c.Packet.(*clientpacket.WalkRequest)
	c.NetState.Send(&serverpacket.MoveAcknowledge{
		Sequence:  p.Sequence,
		Notoriety: uo.NotorietyInnocent,
	})
	return nil
}
