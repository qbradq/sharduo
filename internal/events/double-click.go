package events

// Common DoubleClick events

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Edit", Edit)
	reg("Mount", Mount)
	reg("OpenBackpack", OpenBackpack)
	reg("OpenBankBox", OpenBankBox)
	reg("OpenContainer", OpenContainer)
	reg("OpenPaperDoll", OpenPaperDoll)
	reg("OpenTeleportGUMP", OpenTeleportGUMP)
	reg("PlayerDoubleClick", PlayerDoubleClick)
	reg("TransferHue", TransferHue)
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
	// TODO ownership
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
	game.GetWorld().Map().SendCliloc(receiver, uo.SpeechNormalRange, 1080021,
		strconv.Itoa(bb.ItemCount()), strconv.Itoa(int(bb.Weight()))) // Bank container has ~1_VAL~ items, ~2_VAL~ stones
}

func TransferHue(receiver, source game.Object, v any) {
	if receiver == nil || source == nil {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return
	}
	sm.NetState().Speech(source, "Target object to set hue %d", receiver.Hue())
	sm.NetState().TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		o := game.GetWorld().Find(tr.TargetObject)
		if o == nil {
			return
		}
		o.SetHue(receiver.Hue())
		game.GetWorld().Update(o)
	})
}

func Edit(receiver, source game.Object, v any) {
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	gumps.Edit(sm, receiver)
}

func OpenTeleportGUMP(receiver, source game.Object, v any) {
	if source == nil {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	if sm.NetState() == nil {
		return
	}
	sm.NetState().GUMP(gumps.New("teleport"), source, nil)
}
