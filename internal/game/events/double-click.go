package events

// Common OnDoubleClick events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	evreg.Add("Mount", Mount)
	evreg.Add("OpenPaperDoll", OpenPaperDoll)
	evreg.Add("OpenContainer", OpenContainer)
	evreg.Add("PlayerDoubleClick", PlayerDoubleClick)
}

// PlayerDoubleClick selects between the open paper doll and dismount actions
// based on the current state of the player mobile.
func PlayerDoubleClick(receiver, source game.Object) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	if sm.NetState() == nil {
		return
	}
	if source.Serial() == uo.SerialSystem {
		// Client is explicitly requesting the paper doll
		sm.NetState().OpenPaperDoll(rm)
		return
	}
	if receiver.Serial() != source.Serial() {
		// Someone other than our player is double-clicking on us, always send
		// the paper doll.
		sm.NetState().OpenPaperDoll(rm)
		return
	}
	mio := rm.EquipmentInSlot(uo.LayerMount)
	if mio == nil {
		// We are not mounted, just send ourselves our paper doll
		sm.NetState().OpenPaperDoll(rm)
		return
	}
	// Dismount
	if !rm.Unequip(mio) {
		return
	}
	mi, ok := mio.(*game.MountItem)
	if !ok {
		return
	}
	m := mi.Mount()
	if m == nil {
		return
	}
	m.SetLocation(rm.Location())
	game.GetWorld().Map().SetNewParent(m, nil)
	game.GetWorld().Remove(mi)
}

// OpenPaperDoll opens the paper doll of the receiver mobile to the source.
func OpenPaperDoll(receiver, source game.Object) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	if sm.NetState() == nil {
		return
	}
	sm.NetState().OpenPaperDoll(rm)
}

// OpenContainer opens this container for the mobile. As an additional
// restriction it checks the Z distance against the uo.ContainerOpen* limits.
func OpenContainer(receiver, source game.Object) {
	rc, ok := receiver.(game.Container)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	dz := rc.Location().Z - sm.Location().Z
	if dz < uo.ContainerOpenLowerLimit || dz > uo.ContainerOpenUpperLimit {
		// TODO send cliloc 500312
		return
	}
	rc.Open(sm)
}

// Mount attempts to mount the source mobile onto the receiver.
func Mount(receiver, source game.Object) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	mio := game.GetWorld().New("MountItem")
	if mio == nil {
		return
	}
	mi, ok := mio.(*game.MountItem)
	if !ok {
		return
	}
	switch rm.Body() {
	case 0xC8:
		mi.SetBaseGraphic(0x3E9F)
	case 0xE2:
		mi.SetBaseGraphic(0x3EA0)
	case 0xE4:
		mi.SetBaseGraphic(0x3EA1)
	case 0xCC:
		mi.SetBaseGraphic(0x3EA2)
	}
	if !sm.Equip(mi) {
		return
	}
	// Remove the mount from the world and attach it to the receiver
	game.GetWorld().Map().SetNewParent(rm, mi)
}
