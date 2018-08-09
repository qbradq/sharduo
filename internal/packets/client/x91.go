package client

import (
	"log"

	"github.com/qbradq/sharduo/internal/common"
	"github.com/qbradq/sharduo/internal/packets/server"
)

func x91(r *PacketReader, s *server.NetState) {
	s.BeginCompression()
	s.AddRole(common.RoleAuthenticated)
	r.Seek(5)
	username := r.GetASCII(30)
	password := r.GetASCII(30)
	log.Printf("Auth attempt: %s/%s", username, password)
	// TODO Move this to mobile service
	s.PacketSender().PacketSend(&server.CharacterList{
		CharacterName: username,
		// Flags: common.ClientFlagSixthCharacterSlot |
		// 	common.ClientFlagSeventhCharacterSlot,
	})
}
