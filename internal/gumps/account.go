package gumps

import (
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("account", 0, func() GUMP {
		return &account{}
	})
}

// account implements a menu that displays information about an account and
// allows individual account action.
type account struct {
	StandardGUMP
	Account *game.Account // The account being managed
	Mobile  *game.Mobile  // The mobile associated to the account
}

// Layout implements the game.GUMP interface.
func (g *account) Layout(target, param any) {
	fn := func(x, y int, r game.Role, s string, id uint32) {
		if g.Account.HasRole(r) {
			g.CheckedReplyButton(x*5, y, 5, 1, uo.HueDefault, s, id)
		} else {
			g.ReplyButton(x*5, y, 5, 1, uo.HueDefault, s, id)
		}
	}
	if g.Account == nil {
		if target != nil {
			if m, ok := target.(*game.Mobile); ok && m.Account != nil {
				g.Account = m.Account
				g.Mobile = m
			}
		}
	}
	if g.Account == nil {
		return
	}
	g.Window(10, 7, "Account "+g.Account.Username, 0, 1)
	fn(0, 0, game.RolePlayer, "Player", 1001)
	fn(1, 0, game.RoleModerator, "Mod", 1002)
	fn(0, 1, game.RoleAdministrator, "Admin", 1003)
	fn(1, 1, game.RoleGameMaster, "GM", 1004)
	fn(0, 2, game.RoleDeveloper, "Dev", 1005)
	if g.Account.Locked {
		g.CheckedReplyButton(5, 2, 5, 1, uo.HueDefault, "Locked", 1)
	} else {
		g.ReplyButton(5, 2, 5, 1, uo.HueDefault, "Locked", 1)
	}
	g.HorizontalBar(0, 3, 10)
	g.Text(0, 4, 2, uo.HueDefault, "Email")
	g.TextEntry(2, 4, 8, uo.HueDefault, g.Account.EmailAddress, 256, 2)
	g.Text(0, 5, 10, uo.HueDefault, "Ban Ends "+g.Account.SuspendedUntil.Format(time.RFC3339))
	g.Text(0, 6, 2, uo.HueDefault, "Ban")
	g.ReplyButton(2, 6, 2, 1, uo.HueDefault, "1d", 3)
	g.ReplyButton(4, 6, 2, 1, uo.HueDefault, "3d", 4)
	g.ReplyButton(6, 6, 2, 1, uo.HueDefault, "INF", 5)
	g.ReplyButton(8, 6, 2, 1, uo.HueDefault, "END", 6)
}

// HandleReply implements the GUMP interface.
func (g *account) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	fn := func() {
		m := g.Mobile
		if m != nil && m.NetState != nil {
			m.NetState.Disconnect()
			m.NetState = nil
		}
	}
	// Data
	g.Account.EmailAddress = p.Text(2)
	// Standard reply handler
	if g.StandardReplyHandler(p) {
		return
	}
	// Buttons
	switch p.Button {
	case 1:
		g.Account.Locked = !g.Account.Locked
	case 3:
		g.Account.SuspendedUntil = game.World.ServerTime().Add(time.Hour * 24 * 1)
		fn()
	case 4:
		g.Account.SuspendedUntil = game.World.ServerTime().Add(time.Hour * 24 * 3)
		fn()
	case 5:
		g.Account.SuspendedUntil = game.World.ServerTime().Add(time.Hour * 24 * 365 * 100)
		fn()
	case 6:
		g.Account.SuspendedUntil = time.Time{}
	case 1001:
		g.Account.Roles ^= game.RolePlayer
	case 1002:
		g.Account.Roles ^= game.RoleModerator
	case 1003:
		g.Account.Roles ^= game.RoleAdministrator
	case 1004:
		g.Account.Roles ^= game.RoleGameMaster
	case 1005:
		g.Account.Roles ^= game.RoleDeveloper
	}
}
