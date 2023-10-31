package uod

import (
	"github.com/qbradq/sharduo/internal/events"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	giHandlers.Add(0x13, handleContextMenuRequest)
	giHandlers.Add(0x15, handleContextMenuSelection)
}

var giHandlers = util.NewRegistry[byte, func(*NetState, clientpacket.GeneralInformationPacket)]("gi-packet-handlers")

func handleGeneralInformation(n *NetState, cp clientpacket.Packet) {
	gip := cp.(clientpacket.GeneralInformationPacket)
	fn, ok := giHandlers.Get(gip.SCID())
	if ok {
		fn(n, gip)
	}
}

func handleContextMenuRequest(n *NetState, cp clientpacket.GeneralInformationPacket) {
	p := cp.(*clientpacket.ContextMenuRequest)
	o := world.Find(p.Serial)
	if o == nil {
		return
	}
	menu := game.BuildContextMenu(o, n.m)
	if menu != nil {
		n.Send((*serverpacket.ContextMenu)(menu))
	}
}

func handleContextMenuSelection(n *NetState, cp clientpacket.GeneralInformationPacket) {
	p := cp.(*clientpacket.ContextMenuSelection)
	o := world.Find(p.Serial)
	if o == nil {
		return
	}
	fn := events.GetEventHandlerByIndex(p.EntryID)
	if fn == nil {
		return
	}
	(*fn)(o, n.m, nil)
}
