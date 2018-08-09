package client

import (
	"github.com/qbradq/sharduo/internal/common"
	"github.com/qbradq/sharduo/internal/packets/server"
)

func x5d(r *PacketReader, s *server.NetState) {
	// The user can only access the one character slot, so this packet is just a trigger
	skin := common.RandomSkinHue()
	body := uint16(400)
	id := common.Serial(0x1337)
	x := uint16(1328)
	y := uint16(1626)
	z := byte(50)
	s.PacketSender().PacketSend(&server.PlayerBody{
		ID:           id,
		Body:         body,
		ServerWidth:  7168,
		ServerHeight: 4092,
		X:            x,
		Y:            y,
		Z:            z,
	})
	s.PacketSender().PacketSend(&server.LoginComplete{})
	s.PacketSender().PacketSend(&server.SetClientMap{
		MapID: 0,
	})
	s.PacketSender().PacketSend(&server.DrawObject{
		ID:   id,
		Body: body,
		Hue:  skin,
		X:    x,
		Y:    y,
		Z:    z,
	})
}
