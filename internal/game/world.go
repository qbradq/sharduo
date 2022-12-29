package game

import (
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
	// Map returns the map the world is using.
	Map() *Map
	// GetItemDefinition returns the uo.StaticDefinition that holds the static
	// data for a given item graphic.
	GetItemDefinition(uo.Graphic) *uo.StaticDefinition
	// Random returns the uo.RandomSource for the world.
	Random() uo.RandomSource
}

var world World

// RegisterWorld registers the world object to the game library.
func RegisterWorld(w World) {
	world = w
}
