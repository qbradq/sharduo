package game

// NetState is the interface the server client's network state object must
// implement to be compatible with this library of game objects.
type NetState interface {
	// SendItem sends an object information packet to the client
	SendItem(Item)
	// RemoveObject sends a delete object packet to the client
	RemoveObject(Object)
	// SendDrawPlayer sends a DrawPlayer packet for the attached mobile if any
	SendDrawPlayer()
	// SendWornItem sends the WornItem packet to the given mobile
	SendWornItem(Wearable, Mobile)
}
