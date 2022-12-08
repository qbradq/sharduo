package clientpacket

func init() {
	giFactory.Ignore(0x0005) // Client screen dimensions
	giFactory.Ignore(0x000b) // Client language
	giFactory.Ignore(0x000f) // Client flags
}

var giFactory = NewPacketFactory("general-information")

func newGeneralInformation(in []byte) Packet {
	scid := int(getuint16(in[0:2]))
	data := in[2:]
	return giFactory.New(scid, data)
}
