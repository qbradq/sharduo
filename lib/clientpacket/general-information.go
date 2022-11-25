package clientpacket

func init() {

}

var giFactory = newFactory("general-information")

// GeneralInformation is sent by the client for many reasons
type GeneralInformation struct {
	Base
	// Subcommand number
	Subcommand int
}

func newGeneralInformation(in []byte) Packet {
	scid := int(getuint16(in[0:2]))
	data := in[2:]
	return giFactory.new(scid, data)
}
