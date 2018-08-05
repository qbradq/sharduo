package client

import (
	"net"

	"github.com/qbradq/sharduo/packets/common"
	"github.com/qbradq/sharduo/packets/server"
)

func xA0(r *common.PacketReader, s server.PacketSender) {
	s.PacketSend(&server.ConnectToGameServer{
		Address: net.IPv4(127, 0, 0, 1),
		Port:    2594,
		Key:     0xBAADF00D,
	})
}
