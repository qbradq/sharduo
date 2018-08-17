package common

import "github.com/qbradq/sharduo/pkg/uo"

// A NetState object represents the state of a client's connection with the server.
// All exported functions are thread-safe.
type NetState interface {
	// SetAccount sets the associated account
	SetAccount(a *Account)
	// Authenticated returns true if the client has already been authenticated with an 0x91 packet
	Authenticated() bool
	// CompressOutput returns true if packets sent to this client should be compressed
	CompressOutput() bool
	// BeginCompression makes all new output packets compressed
	BeginCompression()
	// SendPacket sends a packet object to the client over the network. Returns
	// false if the client's output packet channel is full or if the input packet
	// object cannot be cast to a byte slice.
	SendPacket(p uo.ServerPacket) bool
}
