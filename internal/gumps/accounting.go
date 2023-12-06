package gumps

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("accounting", 0, func() GUMP {
		return &accounting{}
	})
}

// accounting implements an interface for managing the accounts on the server.
type accounting struct {
	StandardGUMP
	accounts []*game.Account
}

// Layout implements the game.GUMP interface.
func (g *accounting) Layout(target, param game.Object) {
	g.accounts = game.GetWorld().Accounts()
	pages := len(g.accounts) / 20
	if len(g.accounts)%20 != 0 {
		pages++
	}
	g.Window(10, 20, "Accounting", 0, uint32(pages))
	for i := int(g.currentPage-1) * 20; i < len(g.accounts) && i < int(g.currentPage)*20; i++ {
		a := g.accounts[i]
		ty := i % 20
		g.ReplyButton(0, ty, 6, 1, uo.HueDefault, a.Username(), uint32(1001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *accounting) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	if p.Button >= 1001 {
		// Account button
		i := int(p.Button - 1001)
		if i >= len(g.accounts) {
			return
		}
		a := g.accounts[i]
		agi := New("account")
		ag := agi.(*account)
		ag.Account = a
		n.GUMP(ag, nil, nil)
	}
}
