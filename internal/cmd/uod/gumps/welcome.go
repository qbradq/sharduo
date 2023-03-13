package gumps

import (
	"log"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Welcome implements the standard welcome GUMP.
type Welcome struct {
	game.StandardGUMP
	email string
}

// Layout implements the game.GUMP interface.
func (g *Welcome) Layout(target, param game.Object) {
	tm, ok := target.(game.Mobile)
	if !ok {
		return
	}
	motd, err := data.FS.ReadFile("html/motd.html")
	if err != nil {
		log.Println(err)
	}
	email := ""
	if tm.NetState() != nil {
		email = tm.NetState().Account().EmailAddress()
	}
	if email == "" {
		email = "example@email.com"
	}
	g.Window(12, 16, "Welcome to Trammel Time!", 0)
	g.Page(1)
	g.HTML(0, 0, 12, 12, game.MungHTMLForGUMP(string(motd)), true)
	g.Text(0, 12, 12, uo.HueDefault, "Please provide your email address below. It will only be")
	g.Text(0, 13, 12, uo.HueDefault, "used for password recovery and account information.")
	g.TextEntry(0, 15, 10, uo.HueDefault, email, 128, 3001)
	g.GemButton(10, 15, game.SGGemButtonApply, 1001)
}

// HandleReply implements the GUMP interface.
func (g *Welcome) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	g.email = p.Text(3001)
	if p.Button == 1001 {
		n.Account().SetEmailAddress(g.email)
	}
}
