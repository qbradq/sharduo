package game

import "github.com/qbradq/sharduo/lib/util"

// ContainerObserver is implemented anything that can be notified of changes to
// the contents of a container.
type ContainerObserver interface {
	util.Serialer
	// ContainerOpen sends the open container gump packet and all of the
	// contents of the container to the observer. The observer should keep track
	// of all open containers so they can close then when needed.
	ContainerOpen(Container)
	// ContainerClose closes the container gump on the client side.
	ContainerClose(Container)
	// ContainerItemAdded notifies the observer of a new item in the container.
	ContainerItemAdded(Container, Item)
	// ContainerItemRemoved notifies the observer of a item being removed from
	// the container.
	ContainerItemRemoved(Container, Item)
}
