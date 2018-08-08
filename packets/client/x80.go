package client

import (
	"github.com/qbradq/sharduo/accounting"
	"github.com/qbradq/sharduo/common"
)

func x80(r *common.PacketReader, s common.NetState) {
	r.Seek(1)
	accounting.ServiceRequests <- &accounting.LoginRequest{
		State:    s,
		Username: r.GetASCII(30),
		Password: r.GetASCII(30),
	}
}
