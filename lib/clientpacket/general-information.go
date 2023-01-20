package clientpacket

func init() {
	giFactory.Ignore(0x05) // Client screen dimensions
	giFactory.Ignore(0x0B) // Client language
	giFactory.Ignore(0x0F) // Client flags
}

var giFactory = &packetFactory{}

func newGeneralInformation(in []byte) Packet {
	scid := in[1] // This field is two bytes long but never uses most significate byte
	data := in[2:]
	return giFactory.New(scid, data)
}
