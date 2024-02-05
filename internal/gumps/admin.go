package gumps

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("admin", 0, func() GUMP {
		return &admin{}
	})
}

// admin implements a menu to administer the server
type admin struct {
	StandardGUMP
	shutdownTimer uo.Serial // Shutdown timer serial, if any
}

// Layout implements the game.GUMP interface.
func (g *admin) Layout(target, param any) {
	g.Window(6, 5, "Admin Menu", 0, 1)
	ty := 0
	g.ReplyButton(0, ty, 6, 1, uo.HueDefault, "Accounts", 1)
	ty++
	g.HorizontalBar(0, ty, 6)
	ty++
	g.ReplyButton(0, ty, 6, 1, uo.HueDefault, "Force Save", 2)
	ty++
	g.ReplyButton(0, ty, 6, 1, uo.HueDefault, "Graceful Restart", 3)
	ty++
	g.ReplyButton(0, ty, 6, 1, uo.HueDefault, "Cancel Restart", 4)
	ty++
}

// HandleReply implements the GUMP interface.
func (g *admin) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	switch p.Button {
	case 1:
		n.GUMP(New("accounting"), 0, 0)
	case 2:
		executeCommand(n, "save")
	case 3:
		// Spam the restart message so you can actually see it
		for i := 0; i < 10; i++ {
			executeCommand(n, "broadcast The server will be restarting in 1 minute!")
		}
		g.shutdownTimer = game.NewTimer(uo.DurationMinute*1, "ServerShutdown", nil, n.Mobile(), true, nil)
	case 4:
		if g.shutdownTimer == uo.SerialZero {
			break
		}
		// Spam the restart message so you can actually see it
		for i := 0; i < 10; i++ {
			executeCommand(n, "broadcast Server restart aborted.")
		}
		game.CancelTimer(g.shutdownTimer)
	}
}
