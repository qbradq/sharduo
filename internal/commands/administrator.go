package commands

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Holds administrator-level commands

func init() {
	regcmd(&cmdesc{"location", nil, commandLocation, game.RoleAdministrator, "location", "Tells the absolute location of the targeted location or object"})
	regcmd(&cmdesc{"save", nil, commandSave, game.RoleAdministrator, "save", "Executes a game.GetWorld() save immediately"})
	regcmd(&cmdesc{"shutdown", nil, commandShutdown, game.RoleAdministrator, "shutdown", "Shuts down the server immediately"})
}

func commandLocation(n game.NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeLocation, func(r *clientpacket.TargetResponse) {
		n.Speech(n.Mobile(), "Location X=%d Y=%d Z=%d", r.Location.X, r.Location.Y, r.Location.Z)
	})
}

func commandSave(n game.NetState, args CommandArgs, cl string) {
	saveWorld()
}

func commandShutdown(n game.NetState, args CommandArgs, cl string) {
	shutdown()
}
