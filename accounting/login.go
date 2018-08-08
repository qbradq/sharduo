package accounting

import (
	"log"

	"github.com/qbradq/sharduo/common"
	"github.com/qbradq/sharduo/packets/server"
)

// A LoginRequest requests account authentication
type LoginRequest struct {
	State    common.NetState
	Username string
	Password string
}

func doLogin(r *LoginRequest) {
	log.Println("Login request for", r.Username)
	r.State.PacketSender().PacketSend(&server.GameServerList{
		Name: "ShardUO TC",
	})
}
