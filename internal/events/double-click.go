package events

// Common DoubleClick events

import (
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Edit", Edit)
	reg("HarvestCrop", HarvestCrop)
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
func PlayerDoubleClick(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if receiver.Serial() != source.Serial() {
		// Someone is trying to open our paper doll, just send it
		if sm.NetState() != nil {
			sm.NetState().OpenPaperDoll(rm)
		}
		return true
	}
	if rm.IsMounted() {
		rm.Dismount()
	} else {
		sm.NetState().OpenPaperDoll(rm)
	}
	return true
}

// OpenPaperDoll opens the paper doll of the receiver mobile to the source.
func OpenPaperDoll(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if sm.NetState() != nil {
		sm.NetState().OpenPaperDoll(rm)
	}
	return true
}

// OpenContainer opens this container for the mobile. As an additional
// restriction it checks the Z distance against the uo.ContainerOpen* limits.
func OpenContainer(receiver, source game.Object, v any) bool {
	rc, ok := receiver.(game.Container)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	rl := game.RootParent(receiver).Location()
	sl := game.RootParent(source).Location()
	dz := rl.Z - sl.Z
	if game.RootParent(rc).Location().XYDistance(sm.Location()) > uo.MaxUseRange {
		sm.NetState().Cliloc(nil, 500312) // You cannot reach that.
		return false
	}
	if dz < uo.ContainerOpenLowerLimit || dz > uo.ContainerOpenUpperLimit {
		sm.NetState().Cliloc(nil, 500312) // You cannot reach that.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(rc) {
		sm.NetState().Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	rc.Open(sm)
	return true
}

// Mount attempts to mount the source mobile onto the receiver.
func Mount(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if rm.ControlMaster() == nil || rm.ControlMaster().Serial() != sm.Serial() {
		return false
	}
	// Range check
	if game.RootParent(sm).Location().XYDistance(rm.Location()) > uo.MaxUseRange {
		sm.NetState().Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(rm) {
		sm.NetState().Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	sm.Mount(rm)
	return true
}

// OpenBackpack attempts to open the backpack of the receiver as in snooping or
// pack animals.
func OpenBackpack(receiver, source game.Object, v any) bool {
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if sm.NetState() == nil {
		return false
	}
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	// Range check
	if game.RootParent(sm).Location().XYDistance(rm.Location()) > uo.MaxUseRange {
		sm.NetState().Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(rm) {
		sm.NetState().Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	bpo := rm.EquipmentInSlot(uo.LayerBackpack)
	if bpo == nil {
		// Something very wrong
		return false
	}
	bp, ok := bpo.(game.Container)
	if !ok {
		// Something very wrong
		return false
	}
	bp.Open(sm)
	return true
}

// OpenBankBox attempts to open the bank box of the source.
func OpenBankBox(receiver, source game.Object, v any) bool {
	m, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if m.NetState() == nil {
		return false
	}
	bbo := m.EquipmentInSlot(uo.LayerBankBox)
	if bbo == nil {
		return false
	}
	bb, ok := bbo.(game.Container)
	if !ok {
		return false
	}
	bb.Open(m)
	game.GetWorld().Map().SendCliloc(receiver, uo.SpeechNormalRange, 1080021,
		strconv.Itoa(bb.ItemCount()), strconv.Itoa(int(bb.Weight()))) // Bank container has ~1_VAL~ items, ~2_VAL~ stones
	return true
}

func TransferHue(receiver, source game.Object, v any) bool {
	if receiver == nil || source == nil {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
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
	return true
}

func Edit(receiver, source game.Object, v any) bool {
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil || !sm.NetState().Account().HasRole(game.RoleGameMaster) {
		return false
	}
	gumps.Edit(sm, receiver)
	return true
}

func OpenTeleportGUMP(receiver, source game.Object, v any) bool {
	if source == nil {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if sm.NetState() == nil {
		return false
	}
	sm.NetState().GUMP(gumps.New("teleport"), source, nil)
	return true
}

func HarvestCrop(receiver, source game.Object, v any) bool {
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	// Range check
	if game.RootParent(sm).Location().XYDistance(receiver.Location()) > uo.MaxUseRange {
		sm.NetState().Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(receiver) {
		sm.NetState().Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	tn := strings.TrimSuffix(receiver.TemplateName(), "Crop")
	i := template.Create[game.Item](tn)
	if i == nil {
		return false
	}
	if !sm.DropToBackpack(i, false) {
		sm.DropToFeet(i)
	}
	game.Remove(receiver)
	return true
}
