package uod

import (
	"log"

	"github.com/qbradq/sharduo/lib/clientpacket"
)

func xBD(n *NetState, p clientpacket.Packet) {
	if p, ok := p.(*clientpacket.Version); ok {
		log.Println("Got version packet", p)
	}
}
