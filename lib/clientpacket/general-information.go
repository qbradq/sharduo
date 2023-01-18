package clientpacket

import (
	. "github.com/qbradq/sharduo/lib/dataconv"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	giFactory.Ignore(0x0005) // Client screen dimensions
	giFactory.Ignore(0x000b) // Client language
	giFactory.Ignore(0x000f) // Client flags
}

var giFactory = &packetFactory{}

func newGeneralInformation(in []byte) Packet {
	scid := uo.Serial(GetUint16(in[0:2]))
	data := in[2:]
	return giFactory.New(scid, data)
}
