package game

import (
	"time"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// NetState is implemented by the game server and is responsible for
// communicating with the client.
type NetState interface {
	// Send sends an arbitrary server packet to the connected client. Returns
	// true on success. This function will fail if the connection's send packet
	// channel is full but will not block.
	Send(serverpacket.Packet) bool
	// Disconnect disconnects the backing connection, cleans up the net state
	// and schedules the player's character - if any - to logout.
	Disconnect()
	// SendMobile sends an initial information packet for the mobile.
	SendMobile(*Mobile)
	// SendItem sends an initial information packet for the item.
	SendItem(*Item)
	// UpdateMobile sends an update packet for the mobile.
	UpdateMobile(*Mobile)
	// UpdateItem sends an update packet for the item.
	UpdateItem(*Item)
	// Speech sends a speech packet to the attached client.
	Speech(any, string, ...any)
	// MoveMobile sends a packet to inform the client that the mobile moved.
	MoveMobile(*Mobile)
	// Cliloc sends a localized client message packet to the attached client.
	Cliloc(any, uo.Cliloc, ...string)
	// RemoveMobile sends a packet to the client that removes the mobile from
	// the client's view of the game.
	RemoveMobile(*Mobile)
	// RemoveItem sends a packet to the client that removes the item from the
	// client's view of the game.
	RemoveItem(*Item)
	// ContainerRangeCheck checks all observed containers and closes them as
	// needed based on range.
	ContainerRangeCheck()
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

// Spawner implements an interface allowing for the control of object spawning.
type Spawner interface {
}

// World is the interface the server's game world model must implement for the
// internal game objects to work properly.
type WorldInterface interface {
	// Find returns a pointer to the object with the given ID or nil
	Find(uo.Serial) any
	// FindMobile returns the item with the given serial or nil.
	FindMobile(uo.Serial) *Mobile
	// FindItem returns the item with the given serial or nil.
	FindItem(uo.Serial) *Item
	// Delete removes the given object from the world and deletes it from the
	// data stores.
	Delete(any)
	// UpdateItem schedules an update packet for the item. It is safe to update
	// the same object rapidly in succession. No duplicate packets will be sent.
	UpdateItem(*Item)
	// UpdateMobile schedules an update packet for the item. It is safe to
	// update the same object rapidly in succession. No duplicate packets will
	// be sent.
	UpdateMobile(*Mobile)
	// UpdateItemOPLInfo adds the item to the list of items that must have their
	// OPL data updated client-side.
	UpdateItemOPLInfo(*Item)
	// UpdateMobileOPLInfo adds the mobile to the list of mobiles that must have
	// their OPL data updated client-side.
	UpdateMobileOPLInfo(*Mobile)
	// Map returns the map the world is using.
	Map() *Map
	// GetItemDefinition returns the uo.StaticDefinition that holds the static
	// data for a given item graphic.
	GetItemDefinition(uo.Graphic) *uo.StaticDefinition
	// Time returns the current time in the Sossarian universe. This is what
	// timers use to avoid complications with DST, save lag, rollbacks, and
	// downtime.
	Time() uo.Time
	// ServerTime returns the current wall-clock time of the server. This is
	// updated once per tick.
	ServerTime() time.Time
	// BroadcastPacket sends the packet to every net state connected to the
	// game service with an attached mobile.
	BroadcastPacket(serverpacket.Packet)
	// BroadcastMessage sends a system message to every net state with a mobile
	BroadcastMessage(*Mobile, string, ...any)
	// Accounts returns a slice of pointers to the accounts on the server. This
	// should only be used for admin GUMPs and commands.
	Accounts() []*Account
}
