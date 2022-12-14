package game

// NetState is the interface the server client's network state object must
// implement to be compatible with this library of game objects.
type NetState interface {
	// SendItem sends an object information packet to the client
	SendItem(Item)
}