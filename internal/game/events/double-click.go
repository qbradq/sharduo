package events

// Common DoubleClick events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Mount", Mount)
	reg("OpenBackpack", OpenBackpack)
	reg("OpenBankBox", OpenBankBox)
	reg("OpenContainer", OpenContainer)
	reg("OpenPaperDoll", OpenPaperDoll)
	reg("PlayerDoubleClick", PlayerDoubleClick)
}

// PlayerDoubleClick selects between the open paper doll and dismount actions
// based on the identity of the source.
func PlayerDoubleClick(receiver, source game.Object, v any) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	if receiver.Serial() != source.Serial() {
		// Someone is trying to open our paper doll, just send it
		if sm.NetState() == nil {
			return
		}
		sm.NetState().OpenPaperDoll(rm)
		return
	}
	if rm.IsMounted() {
		rm.Dismount()
	} else {
		sm.NetState().OpenPaperDoll(rm)
	}
}

// OpenPaperDoll opens the paper doll of the receiver mobile to the source.
func OpenPaperDoll(receiver, source game.Object, v any) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
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
func OpenContainer(receiver, source game.Object, v any) {
	rc, ok := receiver.(game.Container)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return
	}
	rl := game.RootParent(receiver).Location()
	sl := game.RootParent(source).Location()
	dz := rl.Z - sl.Z
	if dz < uo.ContainerOpenLowerLimit || dz > uo.ContainerOpenUpperLimit {
		sm.NetState().Cliloc(nil, 500312)
		return
	}
	rc.Open(sm)
}

// Mount attempts to mount the source mobile onto the receiver.
func Mount(receiver, source game.Object, v any) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	sm.Mount(rm)
}

// OpenBackpack attempts to open the backpack of the receiver as in snooping or
// pack animals.
func OpenBackpack(receiver, source game.Object, v any) {
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	if sm.NetState() == nil {
		return
	}
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	bpo := rm.EquipmentInSlot(uo.LayerBackpack)
	if bpo == nil {
		return
	}
	bp, ok := bpo.(game.Container)
	if !ok {
		return
	}
	bp.Open(sm)
}

// OpenBankBox attempts to open the bank box of the source.
func OpenBankBox(receiver, source game.Object, v any) {
	m, ok := source.(game.Mobile)
	if !ok {
		return
	}
	if m.NetState() == nil {
		return
	}
	bbo := m.EquipmentInSlot(uo.LayerBankBox)
	if bbo == nil {
		return
	}
	bb, ok := bbo.(game.Container)
	if !ok {
		return
	}
	bb.Open(m)
}
