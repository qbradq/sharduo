package server

import "github.com/qbradq/sharduo/internal/common"

// DrawPlayer packets are used to draw the client's own character on the screen.
// This seems to work just like a packet 0x77 DrawMobile but is only sent to the
// client playing the mobile.
type DrawPlayer struct {
	ID            common.Serial
	Hue           common.Hue
	Body, X, Y    uint16
	Flags, Dir, Z byte
}

// Compile encodes the state of the Packet object using w
func (p *DrawPlayer) Compile(w *PacketWriter) {
	w.PutByte(0x20)
	w.PutUInt32(uint32(p.ID))
	w.PutUInt16(p.Body)
	w.PutByte(0)
	w.PutUInt16(uint16(p.Hue))
	w.PutByte(p.Flags)
	w.PutUInt16(p.X)
	w.PutUInt16(p.Y)
	w.Fill(0, 2)
	w.PutByte(p.Dir)
	w.PutByte(p.Z)
}
