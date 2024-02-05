package commands

import (
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Holds administrator-level commands

func init() {
	reg(&cmDesc{"broadcast", nil, commandBroadcast, game.RoleAdministrator, "broadcast text", "Broadcasts the given text to all connected players"})
	reg(&cmDesc{"location", []string{"loc"}, commandLocation, game.RoleAdministrator, "location", "Tells the absolute location of the targeted location or object"})
	reg(&cmDesc{"save", nil, commandSave, game.RoleAdministrator, "save", "Executes a game.GetWorld() save immediately"})
	reg(&cmDesc{"shutdown", nil, commandShutdown, game.RoleAdministrator, "shutdown", "Shuts down the server immediately"})
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

func commandBroadcast(n game.NetState, args CommandArgs, cl string) {
	parts := strings.SplitN(cl, " ", 2)
	if len(parts) != 2 {
		return
	}
	broadcast(parts[1])
}
