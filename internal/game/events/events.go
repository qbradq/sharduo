package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

// EventHandler is the function signature of event handlers
type EventHandler func(game.Object, game.Object, any)

var evreg *util.Registry[string, *EventHandler] = util.NewRegistry[string, *EventHandler]("events")
var evbr map[*EventHandler]string = make(map[*EventHandler]string)

func reg(name string, fn EventHandler) {
	evreg.Add(name, &fn)
	evbr[&fn] = name
}

// GetEventHandler returns the named event handler or nil if it does not exist
func GetEventHandler(which string) *EventHandler {
	fn, _ := evreg.Get(which)
	return fn
}

// GetEventName returns the name of the event handler
func GetEventName(fn *EventHandler) string {
	return evbr[fn]
}
