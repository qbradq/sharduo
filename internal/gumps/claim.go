package gumps

import (
	"log"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("claim", func() GUMP {
		return &claim{}
	})
}

// Welcome implements the standard welcome GUMP.
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
	sp := tm.NetState().Account().StabledPets()
	g.Window(10, len(sp), "Claim Pets", 0)
	g.Page(1)
	for i, s := range sp {
		pm := game.Find[game.Mobile](s)
		if pm == nil {
			log.Printf("error: account %s referenced stabled pet %s that does not exist",
				tm.NetState().Account().Username(), s.String())
			continue
		}
		g.ReplyButton(0, i, 10, 1, uo.HueDefault, pm.DisplayName(), uint32(1001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *claim) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	sp := n.Account().StabledPets()
	idx := int(p.Button - 1001)
	if idx < 0 || idx >= len(sp) {
		return
	}
	pm := game.Find[game.Mobile](sp[idx])
	if pm == nil {
		log.Printf("error: account %s referenced stabled pet %s that does not exist",
			n.Account().Username(), sp[idx].String())
		return
	}
	if n.Account().RemoveStabledPet(pm) {
		pm.SetLocation(g.tm.Location())
		game.GetWorld().Map().ForceAddObject(pm)
	}
}
