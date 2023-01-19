package clientpacket

func init() {
	giFactory.Ignore(0x0005) // Client screen dimensions
	giFactory.Ignore(0x000b) // Client language
	giFactory.Ignore(0x000f) // Client flags
}

var giFactory = &packetFactory{}

func newGeneralInformation(in []byte) Packet {
	scid := in[0] // This field is two bytes long but never uses the second
	data := in[2:]
	return giFactory.New(scid, data)
}
