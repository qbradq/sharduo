package server

import "github.com/qbradq/sharduo/common"

// DrawObject packets are used to let the client know what a mobile looks like
type DrawObject struct {
	ID                  common.Serial
	Hue                 common.Hue
	Body, X, Y          uint16
	Noto, Flags, Dir, Z byte
}

// Compile encodes the state of the Packet object using w
func (p *DrawObject) Compile(w *common.PacketWriter) {
	w.PutByte(0x78)
	w.PutUInt16(28)
	w.PutUInt32(uint32(p.ID))
	w.PutUInt16(p.Body)
	w.PutUInt16(p.X)
	w.PutUInt16(p.Y)
	w.PutByte(p.Z)
	w.PutByte(p.Dir)
	w.PutUInt16(uint16(p.Hue))
	w.PutByte(p.Flags)
	w.PutByte(p.Noto)
	w.PutUInt32(0)
	w.PutByte(0)
	w.PutUInt32(0)
}
