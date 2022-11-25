package clientpacket

func init() {
	giFactory.ignore(0x0005) // Client screen dimensions
	giFactory.ignore(0x000b) // Client language
	giFactory.ignore(0x000f) // Client flags
}

var giFactory = newFactory("general-information")

func newGeneralInformation(in []byte) Packet {
	scid := int(getuint16(in[0:2]))
	data := in[2:]
	return giFactory.new(scid, data)
}
