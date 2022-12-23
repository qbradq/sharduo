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
	// SendSpeech sends a speech packet
	SendSpeech(Object, string, ...interface{})

	//
	// Item management and updates
	//

	// SendItem sends an object information packet to the client
	SendObject(Object)
	// RemoveObject sends a delete object packet to the client
	RemoveObject(Object)
	// SendDrawPlayer sends a DrawPlayer packet for the attached mobile if any
	SendDrawPlayer()
	// SendUpdateMobile sends an UpdateMobile packet for the given mobile
	SendUpdateMobile(Mobile)
	// SendWornItem sends the WornItem packet to the given mobile
	SendWornItem(Wearable, Mobile)
	// SendDragItem sends a DragItem packet
	SendDragItem(Item, Mobile, uo.Location, Mobile, uo.Location)
}
