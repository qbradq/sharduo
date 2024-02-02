// Package game implements all of the inner glue code for the game. Actual
// game play elements reside in packages [internal/ai], [internal/events] and
// [internal/gumps].
package game

import "github.com/qbradq/sharduo/lib/uo"

// Time must return the current UO time for the world.
var Time func() uo.Time

// Find must return a pointer to the object or nil.
var Find func(uo.Serial) any

// ExecuteEventHandler must execute the named event handler.
var ExecuteEventHandler func(string, any, any, any) bool

// EventIndex must return a unique index number for the given event name.
var EventIndex func(string) uint16

// World is the current world we are simulating.
var World WorldInterface

// MapLocation returns the map location of the item if it directly on the map,
// or the location of the top-most container if the top-most container is on
// the map, or the location of the mobile who is wearing the top-most container.
// This function is useful during range checks.
func MapLocation(i *Item) uo.Point {
	if i.Wearer != nil {
		return i.Wearer.Location
	}
	for {
		if i.Container == nil {
			return i.Location
		}
		i = i.Container
	}
}
