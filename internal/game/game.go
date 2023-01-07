package game

var eventHandlerGetter func(string) *func(Object, Object)
var eventNameGetter func(*func(Object, Object)) string

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
// the given object with the reciever.
func DynamicDispatch(which string, receiver, source Object) {
	fn := receiver.GetEventHandler(which)
	if fn != nil {
		(*fn)(receiver, source)
	}
}

// SetEventHandlerGetter sets the function used to get event handler functions
// by name.
func SetEventHandlerGetter(fn func(string) *func(Object, Object)) {
	eventHandlerGetter = fn
}

// SetEventNameGetter stes the function used to get event handler names.
func SetEventNameGetter(fn func(*func(Object, Object)) string) {
	eventNameGetter = fn
}
