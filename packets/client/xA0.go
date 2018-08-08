package client

import (
	"net"

	"github.com/qbradq/sharduo/common"
	"github.com/qbradq/sharduo/packets/server"
)

func xA0(r *common.PacketReader, s common.NetState) {
	s.PacketSender().PacketSend(&server.ConnectToGameServer{
		Address: net.ParseIP(common.Config.GetString("network.externalIP", "127.0.0.1")),
		Port:    uint16(common.Config.GetInt("network.port", 2593)),
		Key:     0xBAADF00D,
	})
}
