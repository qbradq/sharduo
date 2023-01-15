package game

var eventHandlerGetter func(string) *func(Object, Object)

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
func DynamicDispatch(which string, receiver, source Object) {
	if receiver == nil {
		return
	}
	fn := receiver.GetEventHandler(which)
	if fn != nil {
		(*fn)(receiver, source)
	}
}

// ExecuteEventHandler executes the named event handler with the given receiver
// and source. Both the receiver and source can be nil.
func ExecuteEventHandler(which string, receiver, source Object) {
	fn := eventHandlerGetter(which)
	if fn != nil {
		(*fn)(receiver, source)
	}
}

// SetEventHandlerGetter sets the function used to get event handler functions
// by name.
func SetEventHandlerGetter(fn func(string) *func(Object, Object)) {
	eventHandlerGetter = fn
}
