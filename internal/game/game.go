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
