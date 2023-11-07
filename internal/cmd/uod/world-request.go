package uod

import (
	"fmt"
	"time"

	"github.com/qbradq/sharduo/internal/configuration"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

// WorldRequest is used to send client and system packets to the world's
// goroutine.
type WorldRequest interface {
	// Returns the NetState associated with this request, if any
	GetNetState() *NetState
	// Execute executes the request
	Execute() error
}

// BaseWorldRequest provides the base implementation of WorldRequest except for
// Execute and GetID to force includers to provide their own.
type BaseWorldRequest struct {
	// The net state associated with the command, if any. System commands tend
	// not to have associated net states.
	NetState *NetState
}

// GetNetState implements the WorldRequest interface
func (r *BaseWorldRequest) GetNetState() *NetState {
	return r.NetState
}

// ClientPacketRequest is sent by the NetState for packets that should be
// addressed directly in the world goroutine.
type ClientPacketRequest struct {
	BaseWorldRequest
	// The client or system packet associated with this command.
	Packet clientpacket.Packet
}

// Execute implements the WorldRequest interface
func (r *ClientPacketRequest) Execute() error {
	handler, found := packetHandlers.Get(r.Packet.ID())
	if !found || handler == nil {
		return fmt.Errorf("unhandled packet 0x%02X", r.Packet.ID())
	}
	handler(r.GetNetState(), r.Packet)
	return nil
}

// CharacterLoginRequest is sent by the server accepting a character login
type CharacterLoginRequest struct {
	BaseWorldRequest
}

// Execute implements the WorldRequest interface
func (r *CharacterLoginRequest) Execute() error {
	var player game.Mobile
	// Attempt to load the player
	if r.NetState.account.Player() != uo.SerialMobileNil {
		o := world.Find(r.NetState.account.Player())
		if p, ok := o.(game.Mobile); ok {
			player = p
		}
		// In case the player mobile was in deep storage we try to remove it
		game.GetWorld().Map().RetrieveObject(player.Serial())
	}
	// Create a new character if needed
	if player == nil {
		// New player setup
		tn := "PlayerMobile"
		if r.NetState.account.HasRole(game.RoleSuperUser | game.RoleDeveloper) {
			tn = "DeveloperMobile"
		} else if r.NetState.account.HasRole(game.RoleGameMaster) {
			tn = "GameMasterMobile"
		} else {
			tn = "AdministratorMobile"
		}
		player = template.Create[game.Mobile](tn)
		player.SetLocation(configuration.StartingLocation)
		player.SetFacing(configuration.StartingFacing)
		// TODO Generic player starting equipment - gold, book, candle, dagger
		i := template.Create[game.Item]("GoldCoin")
		i.SetAmount(100)
		player.DropToBackpack(i, true)
		// TODO DEBUG REMOVE Mining alpha test equipment
		w := template.Create[game.Wearable]("Pickaxe")
		if !player.Equip(w) {
			player.DropToBackpack(w, true)
		}
		i = template.Create[game.Item]("IronOre")
		i.SetAmount(5)
		player.DropToBackpack(i, true)
		i = template.Create[game.Item]("IronIngot")
		i.SetAmount(10)
		player.DropToBackpack(i, true)
	}
	world.Map().SetNewParent(player, nil)
	r.NetState.m = player
	r.NetState.account.SetPlayer(player.Serial())
	r.NetState.m.SetNetState(r.NetState)
	Broadcast("Welcome %s to %s!", r.NetState.m.DisplayName(),
		configuration.GameServerName)
	// Send the EnterWorld packet
	facing := r.NetState.m.Facing()
	if r.NetState.m.IsRunning() {
		facing = facing.SetRunningFlag()
	} else {
		facing = facing.StripRunningFlag()
	}
	r.NetState.Send(&serverpacket.EnterWorld{
		Player:   r.NetState.m.Serial(),
		Body:     r.NetState.m.Body(),
		Location: r.NetState.m.Location(),
		Facing:   facing,
		Width:    uo.MapWidth,
		Height:   uo.MapHeight,
	})
	r.NetState.Send(&serverpacket.LoginComplete{})
	r.NetState.Send(&serverpacket.Time{
		Time: time.Now(),
	})
	world.Map().SendEverything(r.NetState.m)
	r.NetState.SendObject(r.NetState.m)
	r.NetState.GUMP(gumps.New("welcome"), r.NetState.m, nil)
	return nil
}

// CharacterLogoutRequest is sent when the client's network connection ends for
// any reason.
type CharacterLogoutRequest struct {
	BaseWorldRequest

	// Mobile is the mobile of the player logging out
	Mobile game.Mobile
}

// Execute implements the WorldRequest interface
func (r *CharacterLogoutRequest) Execute() error {
	game.NewTimer(uo.DurationSecond*10, "PlayerLogout", r.Mobile, nil, false, nil)
	return nil
}
