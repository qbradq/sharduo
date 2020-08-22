package uod

import (
	"errors"

	"github.com/qbradq/sharduo/lib/clientpacket"
)

func xBD(n *NetState, p clientpacket.Packet) {
	if p, ok := p.(*clientpacket.Version); ok {
		if p.String != "5.0.9.1" {
			n.Error("version check", errors.New("Bad client version"))
		}
	}
}
