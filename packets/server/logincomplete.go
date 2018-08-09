package server

// A LoginComplete packet informs the client that it should present the game world
type LoginComplete struct{}

// Compile encodes the state of the Packet object using w
func (p *LoginComplete) Compile(w *PacketWriter) {
	w.PutByte(0x55)
}
