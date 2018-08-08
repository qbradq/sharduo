package server

import (
	"net"

	"github.com/qbradq/sharduo/common"
)

// A ConnectToGameServer packet directs the client to connect to a game server
type ConnectToGameServer struct {
	Address net.IP
	Port    uint16
	Key     uint32
}

// Compile encodes the state of the Packet object using w
func (p *ConnectToGameServer) Compile(w *common.PacketWriter) {
	w.PutByte(0x8c)
	w.PutIP(p.Address)
	w.PutUInt16(p.Port)
	w.PutUInt32(p.Key)
}
