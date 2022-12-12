package game

import "github.com/qbradq/sharduo/lib/uo"

// World is the interface the server's game world model must implement for the
// internal game objects to work properly.
type World interface {
	// Find returns a pointer to the object with the given ID or nil
	Find(uo.Serial) Object
}

var world World

// RegisterWorld registers the world object to the game library.
func RegisterWorld(w World) {
	world = w
}
