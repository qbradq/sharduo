package client

import (
	"github.com/qbradq/sharduo/accounting"
	"github.com/qbradq/sharduo/packets/common"
	"github.com/qbradq/sharduo/packets/server"
)

func x80(r *common.PacketReader, s server.PacketSender) {
	r.Seek(1)
	accounting.ServiceRequests <- &accounting.LoginRequest{
		Client:   s,
		Username: r.GetASCII(30),
		Password: r.GetASCII(30),
	}
}
