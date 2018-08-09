package accounting

import (
	"log"

	"github.com/qbradq/sharduo/internal/packets/server"
)

// A LoginRequest requests account authentication
type LoginRequest struct {
	State    *server.NetState
	Username string
	Password string
}

func doLogin(r *LoginRequest) {
	log.Println("Login request for", r.Username)
	r.State.PacketSender().PacketSend(&server.GameServerList{
		Name: "ShardUO TC",
	})
}
