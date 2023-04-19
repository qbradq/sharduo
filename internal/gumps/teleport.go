package gumps

import (
	"fmt"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
)

func init() {
	reg("teleport", func() GUMP {
		return &teleport{}
	})
}

type teleportDestination struct {
}

type teleportGroup struct {
	Name         string
	Destinations []teleportDestination
}

// teleport implements a simple menu that allows teleportation of game masters.
type teleport struct {
	StandardGUMP
	currentGroup string
}

// Layout implements the GUMP interface.
func (g *teleport) Layout(target, param game.Object) {
	if g.currentGroup == "" {
		g.Window(11, 11, "Global Teleport Menu", 0)
	} else {
		g.Window(11, 11, fmt.Sprintf("Global Teleport Menu - %s", g.currentGroup), 0)
	}
	g.Page(1)
}

// HandleReply implements the GUMP interface.
func (g *teleport) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	// Standard behavior handling
	if g.StandardReplyHandler(p) {
		return
	}
}
