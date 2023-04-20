package game

import (
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// EventHandler is the function signature of event handlers
type EventHandler func(Object, Object, any)

var eventHandlerGetter func(string) *EventHandler
var eventIndexGetter func(string) uint16

// RootParent returns the top-most parent of the object who's parent is the map.
// If this object's parent is the map this object is returned.
func RootParent(o Object) Object {
	if o == nil {
		return nil
	}
	for {
		p := o.Parent()
		if p == nil || p.Serial() == uo.SerialZero {
			// p.Serial() == uo.SerialZero for TheVoid
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

// Remove completely removes the object and all of its children from the game.
// It additionally removes the objects from the world data store. It is safe to
// pass nil to Remove().
func Remove(o Object) {
	if o == nil {
		return
	}
	p := o.Parent()
	if p == nil {
		// If the object is a direct child of the map we have to send an object
		// remove packet to all net states in range.
		for _, m := range world.Map().GetNetStatesInRange(o.Location(), uo.MaxViewRange) {
			m.NetState().RemoveObject(o)
		}
		world.Map().ForceRemoveObject(o)
	} else {
		// If p is a container it will send remove packets to all observers
		p.ForceRemoveObject(o)
	}
	// If this is a container we need to drop observers
	if c, ok := o.(Container); ok {
		c.StopAllObservers()
	}
	o.SetParent(TheVoid)
	world.Delete(o)
	o.RemoveChildren()
}

// ObjectType returns the marshal type code of the object
func ObjectType(o Object) marshal.ObjectType { return o.ObjectType() }

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
