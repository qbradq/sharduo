package client

import (
	"github.com/qbradq/sharduo/accounting"
	"github.com/qbradq/sharduo/packets/server"
)

func x80(r *PacketReader, s *server.NetState) {
	r.Seek(1)
	accounting.ServiceRequests <- &accounting.LoginRequest{
		State:    s,
		Username: r.GetASCII(30),
		Password: r.GetASCII(30),
	}
}
