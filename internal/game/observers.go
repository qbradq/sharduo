package game

// ContainerObserver is implemented by anything that can be notified of changes
// to the contents of a container.
type ContainerObserver interface {
	// ContainerOpen sends the open container gump packet and all of the
	// contents of the container to the observer. The observer should keep track
	// of all open containers so they can close then when needed.
	ContainerOpen(Container)
	// ContainerClose closes the container on the client side as well as all
	// child containers this observer may be observing.
	ContainerClose(Container)
	// ContainerItemAdded notifies the observer of a new item in the container.
	ContainerItemAdded(Container, Item)
	// ContainerItemRemoved notifies the observer of a item being removed from
	// the container.
	ContainerItemRemoved(Container, Item)
	// ContainerRangeCheck asks the observer to close all out-of-range
	// containers.
	ContainerRangeCheck()
	// ContainerIsObserving returns true if the given container is being
	// observed by the observer.
	ContainerIsObserving(Object) bool
}
