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
	motd, err := data.FS.ReadFile("html/motd.html")
	if err != nil {
		log.Println(err)
	}
	g.Window(12, 16, "Welcome to Trammel Time!", 0)
	g.Page(1)
	g.HTML(0, 0, 12, 12, string(motd), false)
	g.TextEntry(0, 14, 12, uo.HueDefault, g.email, 128, 1)
}

// HandleReply implements the GUMP interface.
func (g *Welcome) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	g.email = p.Text(1)
}
