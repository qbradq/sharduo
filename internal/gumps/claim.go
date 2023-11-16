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
	tm game.Mobile // Target mobile
}

// Layout implements the game.GUMP interface.
func (g *claim) Layout(target, param game.Object) {
	tm, ok := target.(game.Mobile)
	if !ok {
		return
	}
	g.tm = tm
	sp := tm.StabledPets()
	g.Window(10, len(sp), "Claim Pets", 0, 1)
	for i, pm := range sp {
		g.ReplyButton(0, i, 10, 1, uo.HueDefault, pm.DisplayName(), uint32(1001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *claim) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	sp := n.Mobile().StabledPets()
	idx := int(p.Button - 1001)
	if idx < 0 || idx >= len(sp) {
		return
	}
	pm := sp[idx]
	if err := n.Mobile().Claim(pm); err != nil {
		err.SendTo(n, n.Mobile())
	} else {
		game.GetWorld().Map().RetrieveObject(pm.Serial())
		pm.SetLocation(g.tm.Location())
		pm.SetControlMaster(n.Mobile())
		pm.SetAI("Follow")
		pm.SetAIGoal(n.Mobile())
		game.GetWorld().Map().ForceAddObject(pm)
	}
}
