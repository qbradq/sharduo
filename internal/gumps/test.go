package gumps

import (
	"log"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("test", 0, func() GUMP {
		g := &test{
			switches: make([]bool, 6),
			email:    "email@domain.com",
		}
		g.switches[1] = true
		g.switches[5] = true
		return g
	})
}

// test is a test GUMP that uses all of the GUMP features I am aware of and
// support.
type test struct {
	StandardGUMP
	switches []bool
	email    string
}

// Layout implements the GUMP interface.
func (g *test) Layout(target, param game.Object) {
	motd, err := data.FS.ReadFile("html/motd.html")
	if err != nil {
		log.Println(err)
	}
	g.Window(24, 18, "Welcome to the Trammel Time test GUMP", 0, 3)
	switch g.currentPage {
	case 1:
		g.Text(5, 2, 4, uo.HueDefault, "Page 1 HTML")
		g.HTML(0, 5, 24, 10, MungHTMLForGUMP(string(motd)), true)
	case 2:
		g.Text(5, 2, 14, uo.HueDefault, "Page 2 Buttons")
		g.Text(2, 4, 10, uo.HueDefault, "Check Switches")
		g.CheckSwitch(2, 5, 10, 1, uo.HueDefault, "Check Switch 1", 1000, g.switches[0])
		g.CheckSwitch(2, 6, 10, 1, uo.HueDefault, "Check Switch 2", 1001, g.switches[1])
		g.CheckSwitch(2, 7, 10, 1, uo.HueDefault, "Check Switch 2", 1002, g.switches[2])
		g.Text(2, 9, 10, uo.HueDefault, "Radio Switches")
		g.Group()
		g.RadioSwitch(2, 10, 10, 1, uo.HueDefault, "Radio Switch 1", 1003, g.switches[3])
		g.RadioSwitch(2, 11, 10, 1, uo.HueDefault, "Radio Switch 2", 1004, g.switches[4])
		g.RadioSwitch(2, 12, 10, 1, uo.HueDefault, "Radio Switch 2", 1005, g.switches[5])
		g.Text(20, 0, 4, uo.HueDefault, "Gem Buttons")
		for i := 0; i < 17; i++ {
			g.GemButton(21, i+1, SGGemButton(i), 2000+uint32(i))
		}
	case 3:
		g.Text(5, 2, 4, uo.HueDefault, "Page 3 Text Entry")
		g.TextEntry(5, 5, 14, uo.HueIce3, g.email, 128, 3000)
	}
}

// HandleReply implements the GUMP interface.
func (g *test) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	n.Speech(nil, "Button reply %d", p.Button)
	for i := uint32(0); i < 6; i++ {
		g.switches[i] = p.Switch(1000 + i)
	}
	g.email = p.Text(3000)
}
