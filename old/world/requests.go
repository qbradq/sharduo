package world

import (
	"math/rand"

	"github.com/qbradq/sharduo/internal/common"
	"github.com/qbradq/sharduo/pkg/uo"
)

// A NewCharacterRequest asks the root instance to create a new mobile for an
// account.
type NewCharacterRequest struct {
	State common.NetState
	Name  string
}

func doNewCharacter(r *NewCharacterRequest) {
	m := &Mobile{
		Object{
			Serial: 1,
			Name:   r.Name,
			Hue:    uo.RandomSkinHue(),
			Body:   uint16(400 + rand.Intn(1)),
			X:      1328,
			Y:      1624,
			Z:      50,
			Dir:    uo.DirWest,
		},
	}
	attachToMobile(r.State, m)
}

func attachToMobile(s common.NetState, m *Mobile) {
	pb := uo.NewServerPacketPlayerBody(make([]byte, 0, 1024),
		m.Serial, m.Body, m.X, m.Y, m.Z, m.Dir, 7168, 4092)
	s.SendPacket(pb)
	lc := uo.NewServerPacketLoginComplete(make([]byte, 0, 8))
	s.SendPacket(lc)
	sm := uo.NewServerPacketSetMap(make([]byte, 0, 8), 0)
	s.SendPacket(sm)
	dm := uo.NewServerPacketDrawMobile(make([]byte, 0, 1024),
		m.Serial, m.Body, m.X, m.Y, m.Z, m.Dir, m.Hue, uo.StatusNormal, uo.NotoFriend)
	dm.Finish()
	s.SendPacket(dm)
	dp := uo.NewServerPacketDrawPlayer(make([]byte, 0, 1024),
		m.Serial, m.Body, m.Hue, uo.StatusNormal, m.X, m.Y, m.Z, m.Dir)
	s.SendPacket(dp)
}
