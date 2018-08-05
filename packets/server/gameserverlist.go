package server

// A GameServerList packet gives the list of game servers to the client
type GameServerList struct {
	Name string
}

// Compile encodes the state of the Packet object using w
func (p *GameServerList) Compile(w *PacketWriter) {
	w.PutByte(0xa8)
	w.PutUInt16(46)
	w.PutByte(0x5d)
	w.PutUInt16(1)
	w.PutUInt16(0)
	w.PutASCII(p.Name, 32)
	w.PutByte(0)
	w.PutByte(0)
	w.PutByte(127)
	w.PutByte(0)
	w.PutByte(0)
	w.PutByte(1)
}
