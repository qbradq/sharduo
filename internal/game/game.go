package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/util"
)

// EventHandler is the function signature of event handlers
type EventHandler func(Object, Object, any)

var eventHandlerGetter func(string) *EventHandler
var eventIndexGetter func(string) uint16
var templateObjectGetter func(string) *util.TagFileObject

// RootParent returns the top-most parent of the object who's parent is the map.
// If this object's parent is the map this object is returned.
func RootParent(o Object) Object {
	if o == nil {
		return nil
	}
	for {
		if o.Parent() == nil {
			return o
		}
		o = o.Parent()
	}
}

// DynamicDispatch attempts to execute the named dynamic dispatch function on
// the given object with the receiver. The receiver may not be nil, but the
// source can.
func DynamicDispatch(which string, receiver, source Object, v any) {
	if receiver == nil {
		return
	}
	fn := receiver.GetEventHandler(which)
	if fn != nil {
		(*fn)(receiver, source, v)
	}
}

// ExecuteEventHandler executes the named event handler with the given receiver
// and source. Both the receiver and source can be nil.
func ExecuteEventHandler(which string, receiver, source Object, v any) {
	fn := eventHandlerGetter(which)
	if fn != nil {
		(*fn)(receiver, source, v)
	}
}

// BuildContextMenu builds the context menu for the given object.
func BuildContextMenu(o Object) *ContextMenu {
	p := &ContextMenu{}
	(*serverpacket.ContextMenu)(p).Serial = o.Serial()
	o.AppendContextMenuEntries(p)
	return p
}

// DispatchDeserialize executes the Deserialize method through the Object
// interface.
func DispatchDeserialize(o Object, tfo *util.TagFileObject) {
	o.Deserialize(tfo)
}

// SetEventHandlerGetter sets the function used to get event handler functions
// by name.
func SetEventHandlerGetter(fn func(string) *EventHandler) {
	eventHandlerGetter = fn
}

// SetEventIndexGetter sets the function used to get event handler index numbers
// by name.
func SetEventIndexGetter(fn func(string) uint16) {
	eventIndexGetter = fn
}

// SetTemplateObjectGetter sets the function used to get template objects by
// name.
func SetTemplateObjectGetter(fn func(string) *util.TagFileObject) {
	templateObjectGetter = fn
}
