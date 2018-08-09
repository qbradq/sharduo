package server

// A CharacterList packet displays the character selection list to the client
type CharacterList struct {
	CharacterName string
	Flags         FeatureFlag
}

// Compile encodes the state of the Packet object using w
func (p *CharacterList) Compile(w *PacketWriter) {
	w.PutByte(0xa9)
	w.PutUInt16(372)
	w.PutByte(1)

	w.PutASCII(p.CharacterName, 30)
	w.Fill(0, 30+60*4)
	w.PutByte(1)
	w.PutByte(0)
	w.PutASCII("Serpent's Hold", 31)
	w.PutASCII("The Order Hall", 31)
	w.PutUInt32(uint32(p.Flags))
}
