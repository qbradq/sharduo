package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	game.ExecuteEventHandler = func(e string, r, s, p any) bool {
		fn := GetEventHandler(e)
		if fn == nil {
			return true
		}
		return (*fn)(r, s, p)
	}
}

// EventHandler is the function signature of event handlers
type EventHandler func(any, any, any) bool

// Event handler registrar
var evreg *util.Registry[string, *EventHandler] = util.NewRegistry[string, *EventHandler]("events")

// Event handler back-reference registrar
var evbr map[*EventHandler]string = map[*EventHandler]string{}

// Event handler index back-reference registrar
var evhibr map[*EventHandler]uint16 = map[*EventHandler]uint16{}

// Event handlers by index
var evidx []*EventHandler = []*EventHandler{nil}

func reg(name string, fn EventHandler) {
	evreg.Add(name, &fn)
	evbr[&fn] = name
	idx := uint16(len(evidx))
	evidx = append(evidx, &fn)
	evhibr[&fn] = idx
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

// GetEventIndex returns the index number of the event handler. A return value
// of 0 means nil or event handler not found.
// NOTE: This index number can change when events are added. DO NOT PERSIST THIS
// VALUE!
func GetEventIndex(fn *EventHandler) uint16 {
	return evhibr[fn]
}

// GetEventHandlerByIndex returns the event handler by index or nil if it does
// not exist.
func GetEventHandlerByIndex(idx uint16) *EventHandler {
	if idx >= uint16(len(evidx)) {
		return nil
	}
	return evidx[idx]
}

// GetEventHandlerIndex returns the event handler index by name. A return value
// of 0 means nil or event handler not found.
// NOTE: This index number can change when events are added. DO NOT PERSIST THIS
// VALUE!
func GetEventHandlerIndex(which string) uint16 {
	fn := GetEventHandler(which)
	if fn == nil {
		return 0
	}
	return evhibr[fn]
}
