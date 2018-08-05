package server

import "net"

// A ConnectToGameServer packet directs the client to connect to a game server
type ConnectToGameServer struct {
	Address net.IP
	Port    uint16
	Key     uint32
}

// Compile encodes the state of the Packet object using w
func (p *ConnectToGameServer) Compile(w *PacketWriter) {
	w.PutByte(0x8c)
	w.PutIP(p.Address)
	w.PutUInt16(p.Port)
	w.PutUInt32(p.Key)
}
