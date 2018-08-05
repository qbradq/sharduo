package accounting

import (
	"log"

	"github.com/qbradq/sharduo/packets/server"
)

// A LoginRequest requests account authentication
type LoginRequest struct {
	Client   server.PacketSender
	Username string
	Password string
}

func doLogin(r *LoginRequest) {
	log.Println("Login request for", r.Username)
	r.Client.PacketSend(&server.LoginDenied{
		Reason: server.LoginDeniedReasonAccountBlocked,
	})
}
