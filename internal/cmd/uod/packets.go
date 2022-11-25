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
	clientPacketFactory.add(0x34, ignorePacket)
	clientPacketFactory.add(0x73, handleClientPing)
	clientPacketFactory.add(0xad, handleClientSpeech)
	clientPacketFactory.add(0xbd, handleClientVersion)
	clientPacketFactory.add(0xc8, handleClientViewRange)
}

var clientPacketFactory = newPacketHandlerFactory("client")

func ignorePacket(n *NetState, cp clientpacket.Packet) {
	// Do nothing
}

func handleClientPing(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Ping)
	n.Send(&serverpacket.Ping{
		Key: p.Key,
	})
}

func handleClientSpeech(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Speech)
	if n.m != nil {
		GlobalChat(n.m.Name, p.Text)
	}
}

func handleClientVersion(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Version)
	if p.String != "7.0.15.1" {
		n.Error("version check", errors.New("bad client version"))
	}
}

func handleClientViewRange(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.ClientViewRange)
	n.Send(&serverpacket.ClientViewRange{
		Range: byte(p.Range),
	})
}

func handleWalkRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.WalkRequest)
	n.Send(&serverpacket.MoveAcknowledge{
		Sequence:  p.Sequence,
		Notoriety: uo.NotorietyInnocent,
	})
}
