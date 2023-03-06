package gumps

import (
	"log"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Test is a test GUMP that uses all of the GUMP features I am aware of and
// support.
type Test struct {
	game.StandardGUMP
}

// Layout implements the GUMP interface.
func (g *Test) Layout(target, param game.Object) {
	motd, err := data.FS.ReadFile("html/motd.html")
	if err != nil {
		log.Println(err)
	}
	g.Window(24, 18, "Welcome to the Trammel Time test GUMP", 0)
	g.Page(1)
	g.Text(5, 2, 4, 1, uo.HueDefault, "Page 1")
	g.HTML(0, 5, 24, 10, string(motd), true)
	g.Page(2)
	g.Text(5, 2, 4, 1, uo.HueDefault, "Page 2")
	g.Page(3)
	g.Text(5, 2, 4, 1, uo.HueDefault, "Page 3")
	g.Page(4)
	g.Text(5, 2, 4, 1, uo.HueDefault, "Page 4")
}

// HandleReply implements the GUMP interface.
func (g *Test) HandleReply(p *clientpacket.GUMPReply) {
}
