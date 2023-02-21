package clientpacket

import (
	dc "github.com/qbradq/sharduo/lib/dataconv"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	giFactory.Ignore(0x05) // Client screen dimensions
	giFactory.Ignore(0x0B) // Client language
	giFactory.Ignore(0x0F) // Client flags
	giFactory.Add(0x15, newContextMenuSelection)
}

var giFactory = &packetFactory{}

type GeneralInformationPacket interface {
	ID() byte
	SCID() byte
}

type baseGIPacket struct {
	id byte
	sc byte
}

func (p *baseGIPacket) ID() byte { return p.id }

func (p *baseGIPacket) SCID() byte { return p.sc }

func newGeneralInformation(in []byte) Packet {
	scid := in[1] // This field is two bytes long but never uses the most significate byte
	data := in[2:]
	return giFactory.New(scid, data)
}

// ContextMenuSelection is sent by the client when the user selects a context
// menu entry.
type ContextMenuSelection struct {
	baseGIPacket
	// Serial of the object the context menu was opened on
	Serial uo.Serial
	// Unique ID of the entry selected
	EntryID uint16
}

func newContextMenuSelection(in []byte) Packet {
	return &ContextMenuSelection{
		baseGIPacket: baseGIPacket{id: 0xBF, sc: 0x15},
		Serial:       uo.Serial(dc.GetUint32(in[0:4])),
		EntryID:      dc.GetUint16(in[4:6]),
	}
}
