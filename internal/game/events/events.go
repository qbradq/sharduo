package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

var evreg *util.Registry[string, *func(game.Object, game.Object)] = util.NewRegistry[string, *func(game.Object, game.Object)]("events")
var evbr map[*func(game.Object, game.Object)]string = make(map[*func(game.Object, game.Object)]string)

func reg(name string, fn func(game.Object, game.Object)) {
	evreg.Add(name, &fn)
	evbr[&fn] = name
}

// GetEventHandler returns the named event handler or nil if it does not exist
func GetEventHandler(which string) *func(game.Object, game.Object) {
	fn, _ := evreg.Get(which)
	return fn
}

// GetEventName returns the name of the event handler
func GetEventName(fn *func(game.Object, game.Object)) string {
	return evbr[fn]
}
