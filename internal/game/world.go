package game

import (
	"time"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// World is the interface the server's game world model must implement for the
// internal game objects to work properly.
type World interface {
	// New returns a pointer to a newly constructed object that has been
	// assigned a unique serial of the correct type and added to the data
	// stores. It does not yet have a parent nor has it been placed on the map.
	// One of these things must be done otherwise the data store will leak this
	// object.
	New(string) Object
	// Find returns a pointer to the object with the given ID or nil
	Find(uo.Serial) Object
	// Remove removes the given object from the world, removing it from the
	// current parent forcefully and deleting it from the data stores.
	Remove(Object)
	// Update adds the object to the world's list of objects that we need to
	// send update packets for. It is safe to update the same object rapidly in
	// succession. No duplicate packets will be sent.
	Update(Object)
	// Map returns the map the world is using.
	Map() *Map
	// GetItemDefinition returns the uo.StaticDefinition that holds the static
	// data for a given item graphic.
	GetItemDefinition(uo.Graphic) *uo.StaticDefinition
	// Random returns the uo.RandomSource for the world.
	Random() uo.RandomSource
	// Time returns the current time in the Sossarian universe. This is what
	// timers use to avoid complications with DST, save lag, rollbacks, and
	// downtime.
	Time() uo.Time
	// ServerTime returns the current wall-clock time of the server. This is
	// updated once per tick.
	ServerTime() time.Time
	// BroadcastPacket sends the packet to every net state connected to the
	// game service with an attached mobile.
	BroadcastPacket(serverpacket.Packet)
	// BroadcastMessage sends a system message to every net state with a mobile
	BroadcastMessage(Object, string, ...interface{})
}

var world World

// RegisterWorld registers the world object to the game library.
func RegisterWorld(w World) {
	world = w
}

// GetWorld returns the internal world object.
func GetWorld() World {
	return world
}
