// Package game implements all of the inner glue code for the game. Actual
// game play elements reside in packages [internal/ai], [internal/events] and
// [internal/gumps].
package game

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

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

// UOError represents a game rules violation and contains information on how
// to alert the player.
type UOError struct {
	Cliloc    uo.Cliloc // Cliloc of the error message or zero if the error message is a string
	Arguments []string  // Arguments to Cliloc, only valid if Cliloc is non zero
	Message   string    // String message, only valid if Cliloc is zero
}

// Error implements the Error interface.
func (e *UOError) Error() string {
	if e.Cliloc != 0 {
		return fmt.Sprintf("cliloc error %d %v", e.Cliloc, e.Arguments)
	}
	return e.Message
}

// Packet returns the server packet to send for this error.
func (e *UOError) Packet() serverpacket.Packet {
	if e.Cliloc != 0 {
		return &serverpacket.ClilocMessage{
			Hue:       1153,
			Cliloc:    e.Cliloc,
			Arguments: e.Arguments,
		}
	}
	return &serverpacket.Speech{
		Hue:  1153,
		Type: uo.SpeechTypeNormal,
		Text: e.Message,
	}
}
