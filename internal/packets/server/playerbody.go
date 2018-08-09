package server

import "github.com/qbradq/sharduo/pkg/uo"

// A PlayerBody packet is sent after login and presumably after crossing server
// boundaries on OSI servers. This must be sent at least once prior to the
// LoginComplete packet or the player will get a random view of the map.
type PlayerBody struct {
	ID                                    uo.Serial
	Body, X, Y, ServerWidth, ServerHeight uint16
	Dir, Z                                byte
}

// Compile encodes the state of the Packet object using w
func (p *PlayerBody) Compile(w *PacketWriter) {
	w.PutByte(0x1b)
	w.PutUInt32(uint32(p.ID))
	w.Fill(0, 4)
	w.PutUInt16(p.Body)
	w.PutUInt16(p.X)
	w.PutUInt16(p.Y)
	w.PutByte(0)
	w.PutByte(p.Z)
	w.PutByte(p.Dir)
	w.Fill(0, 9)
	w.PutUInt16(p.ServerWidth - 8)
	w.PutUInt16(p.ServerHeight)
	w.Fill(0, 6)
}
