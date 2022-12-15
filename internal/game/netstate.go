package game

// NetState is the interface the server client's network state object must
// implement to be compatible with this library of game objects.
type NetState interface {
	// ViewRange returns the current view range of the client connection
	ViewRange() int
	// SendItem sends an object information packet to the client
	SendItem(Item)
	// RemoveObject sends a delete object packet to the client
	RemoveObject(Object)
}
