package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	game.ExecuteEventHandler = ExecuteEventHandler
	game.EventIndex = getEventHandlerIndex
}

// eventHandler is the function signature of event handlers
type eventHandler func(any, any, any) bool

// Event handler registrar
var evReg *util.Registry[string, *eventHandler] = util.NewRegistry[string, *eventHandler]("events")

// Event handler back-reference registrar
var evBr map[*eventHandler]string = map[*eventHandler]string{}

// Event handler index back-reference registrar
var ehIBR map[*eventHandler]uint16 = map[*eventHandler]uint16{}

// Event handlers by index
var evIdx []*eventHandler = []*eventHandler{nil}

func reg(name string, fn eventHandler) {
	evReg.Add(name, &fn)
	evBr[&fn] = name
	idx := uint16(len(evIdx))
	evIdx = append(evIdx, &fn)
	ehIBR[&fn] = idx
}

// getEventHandler returns the named event handler or nil if it does not exist
func getEventHandler(which string) *eventHandler {
	fn, _ := evReg.Get(which)
	return fn
}

// getEventName returns the name of the event handler
func getEventName(fn *eventHandler) string {
	return evBr[fn]
}

// getEventIndex returns the index number of the event handler. A return value
// of 0 means nil or event handler not found.
// NOTE: This index number can change when events are added. DO NOT PERSIST THIS
// VALUE!
func getEventIndex(fn *eventHandler) uint16 {
	return ehIBR[fn]
}

// GetEventHandlerByIndex returns the event handler by index or nil if it does
// not exist.
func GetEventHandlerByIndex(idx uint16) *eventHandler {
	if idx >= uint16(len(evIdx)) {
		return nil
	}
	return evIdx[idx]
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
	return ehIBR[fn]
}

// ExecuteEventHandler executes the named event handler with the parameters.
func ExecuteEventHandler(e string, r, s, p any) bool {
	fn := getEventHandler(e)
	if fn == nil {
		return true
	}
	return (*fn)(r, s, p)
}
