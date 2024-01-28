package game

// NetState is implemented by the game server and is responsible for
// communicating with the client.
type NetState interface {
}

// ContainerObserver is implemented by anything that can be notified of changes
// to the contents of a container.
type ContainerObserver interface {
	// ContainerOpen sends the open container gump packet and all of the
	// contents of the container to the observer. The observer should keep track
	// of all open containers so they can close then when needed.
	ContainerOpen(*Item)
	// ContainerClose closes the container on the client side as well as all
	// child containers this observer may be observing.
	ContainerClose(*Item)
	// ContainerItemAdded notifies the observer of a new item in the container.
	ContainerItemAdded(*Item, *Item)
	// ContainerItemRemoved notifies the observer of a item being removed from
	// the container.
	ContainerItemRemoved(*Item, *Item)
	// ContainerItemOPLChanged notifies the observer of an item's OPL changing.
	ContainerItemOPLChanged(*Item, *Item)
	// ContainerRangeCheck asks the observer to close all out-of-range
	// containers.
	ContainerRangeCheck()
	// ContainerIsObserving returns true if the given container is being
	// observed by the observer.
	ContainerIsObserving(*Item) bool
}
