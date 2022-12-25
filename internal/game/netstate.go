package game

import "github.com/qbradq/sharduo/lib/uo"

// NetState is the interface the server client's network state object must
// implement to be compatible with this library of game objects.
type NetState interface {
	//
	// Speech and messaging
	//

	// SystemMessage sends a system message
	SystemMessage(string, ...interface{})
	// Speech sends a speech packet
	Speech(Object, string, ...interface{})

	//
	// Item management and updates
	//

	// SendItem sends an object information packet to the client
	SendObject(Object)
	// RemoveObject sends a delete object packet to the client
	RemoveObject(Object)
	// DrawPlayer sends a DrawPlayer packet for the attached mobile if any
	DrawPlayer()
	// UpdateMobile sends an UpdateMobile packet for the given mobile
	UpdateMobile(Mobile)
	// WordItem sends the WornItem packet to the given mobile
	WornItem(Wearable, Mobile)
	// DragItem sends a DragItem packet
	DragItem(Item, Mobile, uo.Location, Mobile, uo.Location)
	// DropReject sends the MoveItemReject packet with the given reason code
	DropReject(uo.MoveItemRejectReason)

	//
	// Containers
	//

	// OpenContainer opens a container gump on the client
	OpenContainer(Container)
	// AddItemToContainer adds an item to a container gump on the client
	AddItemToContainer(Container, Item)
	// RemoveItemFromContainer removes an item from a container on the client
	RemoveItemFromContainer(Container, Item)

	//
	// Gumps
	//

	// CloseGump closes the named gump on the client
	CloseGump(gump uo.Serial)
}
