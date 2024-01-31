package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	game.ExecuteEventHandler = func(e string, r, s, p any) bool {
		fn := getEventHandler(e)
		if fn == nil {
			return true
		}
		return (*fn)(r, s, p)
	}
	game.EventIndex = getEventHandlerIndex
}

// eventHandler is the function signature of event handlers
type eventHandler func(any, any, any) bool

// Event handler registrar
var evreg *util.Registry[string, *eventHandler] = util.NewRegistry[string, *eventHandler]("events")

// Event handler back-reference registrar
var evbr map[*eventHandler]string = map[*eventHandler]string{}

// Event handler index back-reference registrar
var evhibr map[*eventHandler]uint16 = map[*eventHandler]uint16{}

// Event handlers by index
var evidx []*eventHandler = []*eventHandler{nil}

func reg(name string, fn eventHandler) {
	evreg.Add(name, &fn)
	evbr[&fn] = name
	idx := uint16(len(evidx))
	evidx = append(evidx, &fn)
	evhibr[&fn] = idx
}

// getEventHandler returns the named event handler or nil if it does not exist
func getEventHandler(which string) *eventHandler {
	fn, _ := evreg.Get(which)
	return fn
}

// getEventName returns the name of the event handler
func getEventName(fn *eventHandler) string {
	return evbr[fn]
}

// getEventIndex returns the index number of the event handler. A return value
// of 0 means nil or event handler not found.
// NOTE: This index number can change when events are added. DO NOT PERSIST THIS
// VALUE!
func getEventIndex(fn *eventHandler) uint16 {
	return evhibr[fn]
}

// GetEventHandlerByIndex returns the event handler by index or nil if it does
// not exist.
func GetEventHandlerByIndex(idx uint16) *eventHandler {
	if idx >= uint16(len(evidx)) {
		return nil
	}
	return evidx[idx]
}

// getEventHandlerIndex returns the event handler index by name. A return value
// of 0 means nil or event handler not found.
// NOTE: This index number can change when events are added. DO NOT PERSIST THIS
// VALUE!
func getEventHandlerIndex(which string) uint16 {
	fn := getEventHandler(which)
	if fn == nil {
		return 0
	}
	return evhibr[fn]
}
