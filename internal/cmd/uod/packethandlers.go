package uod

import (
	"errors"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
)

func x73(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Ping)
	n.Send(&serverpacket.Ping{
		Key: p.Key,
	})
}

func xAD(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Speech)
	if n.m != nil {
		GlobalChat(n.m.Name, p.Text)
	}
}

func xBD(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Version)
	if p.String != "5.0.9.1" {
		n.Error("version check", errors.New("Bad client version"))
	}
}
