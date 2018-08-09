package server

// A SetClientMap packet instructs the client to load and display the numbered map.
type SetClientMap struct {
	MapID byte
}

// Compile encodes the state of the Packet object using w
func (p *SetClientMap) Compile(w *PacketWriter) {
	w.PutByte(0xbf)
	w.PutUInt16(6)
	w.PutUInt16(0x0008)
	w.PutByte(p.MapID)
}
