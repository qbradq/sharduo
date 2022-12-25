package uod

import (
	"fmt"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
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
	handler, found := worldHandlers.Get(r.Packet.Serial())
	if !found || handler == nil {
		return fmt.Errorf("unhandled packet %s", r.Packet.Serial().String())
	}
	handler(r.GetNetState(), r.Packet)
	return nil
}

// SpeechCommandRequest is sent by a player using the speech packet with the
// message starting with the '[' character
type SpeechCommandRequest struct {
	BaseWorldRequest
	Command Command
}

// Execute implements the WorldRequest interface
func (r *SpeechCommandRequest) Execute() error {
	return r.Command.Execute(r.NetState)
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
	}
	// Create a new character if needed
	if player == nil {
		player = world.New("Player").(game.Mobile)
		// TODO New player setup
		world.Map().SetNewParent(player, nil)
	}
	r.NetState.m = player
	r.NetState.account.SetPlayer(player.Serial())
	r.NetState.m.SetNetState(r.NetState)
	Broadcast("Welcome %s to Trammel Time!", r.NetState.m.DisplayName())

	// Send the EnterWorld packet
	facing := r.NetState.m.Facing()
	if r.NetState.m.IsRunning() {
		facing = facing.SetRunningFlag()
	} else {
		facing = facing.StripRunningFlag()
	}
	r.NetState.Send(&serverpacket.EnterWorld{
		Player: r.NetState.m.Serial(),
		Body:   r.NetState.m.Body(),
		X:      r.NetState.m.Location().X,
		Y:      r.NetState.m.Location().Y,
		Z:      r.NetState.m.Location().Z,
		Facing: facing,
		Width:  uo.MapWidth,
		Height: uo.MapHeight,
	})
	r.NetState.Send(&serverpacket.LoginComplete{})
	world.Map().SendEverything(r.NetState.m)
	r.NetState.SendObject(r.NetState.m)

	return nil
}
