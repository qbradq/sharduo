package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

var evreg *util.Registry[string, func(game.Object, game.Object)] = util.NewRegistry[string, func(game.Object, game.Object)]("events")

// GetEventHandler returns the named event handler or nil if it does not exist
func GetEventHandler(which string) func(game.Object, game.Object) {
	fn, _ := evreg.Get(which)
	return fn
}
