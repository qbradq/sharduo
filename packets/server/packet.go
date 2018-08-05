// Package server contains all server packet structures
package server

// A Packet is an object that can encode its state into a byte slice in Ultima
// Online wire protocol format
type Packet interface {
	Compile(buf []byte) []byte
}

// A PacketSender can transmit Packet objects to consumers gracefully and
// without blocking
type PacketSender interface {
	PacketSend(p Packet)
}
