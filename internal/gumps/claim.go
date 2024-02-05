package gumps

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("claim", 0, func() GUMP {
		return &claim{}
	})
}

// claim implements a list that lets players claim stabled pets
type claim struct {
	StandardGUMP
	tm *game.Mobile // Target mobile
}

// Layout implements the game.GUMP interface.
func (g *claim) Layout(target, param any) {
	tm, ok := target.(*game.Mobile)
	if !ok || tm.PlayerData == nil {
		return
	}
	g.tm = tm
	g.Window(10, len(tm.PlayerData.StabledPets), "Claim Pets", 0, 1)
	for i, pm := range tm.PlayerData.StabledPets {
		g.ReplyButton(0, i, 10, 1, uo.HueDefault, pm.DisplayName(), uint32(1001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *claim) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if n.Mobile() == nil || n.Mobile().PlayerData == nil {
		return
	}
	if g.StandardReplyHandler(p) {
		return
	}
	idx := int(p.Button - 1001)
	if idx < 0 || idx >= len(n.Mobile().PlayerData.StabledPets) {
		return
	}
	pm := n.Mobile().PlayerData.StabledPets[idx]
	n.Mobile().PlayerData.Claim(pm, n.Mobile())
}
