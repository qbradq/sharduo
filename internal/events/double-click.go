package events

// Common DoubleClick events

import (
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("CashCheck", cashCheck)
	reg("Edit", editObject)
	reg("HarvestCrop", harvestCrop)
	reg("Mount", mountMobile)
	reg("OpenBackpack", openBackpack)
	reg("OpenBankBox", openBankBox)
	reg("OpenContainer", openContainer)
	reg("OpenPaperDoll", openPaperDoll)
	reg("OpenTeleportGUMP", openTeleportGUMP)
	reg("PlayerDoubleClick", playerDoubleClick)
	reg("TransferHue", transferHue)
}

// playerDoubleClick selects between the open paper doll and dismount actions
// based on the identity of the source.
func playerDoubleClick(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if rm.Serial != sm.Serial {
		// Someone is trying to open our paper doll, just send it
		if sm.NetState != nil {
			sm.NetState.OpenPaperDoll(rm)
		}
		return true
	}
	if rm.Equipment[uo.LayerMount] != nil {
		rm.Dismount()
	} else {
		sm.NetState.OpenPaperDoll(rm)
	}
	return true
}

// openPaperDoll opens the paper doll of the receiver mobile to the source.
func openPaperDoll(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if sm.NetState != nil {
		sm.NetState.OpenPaperDoll(rm)
	}
	return true
}

// openContainer opens this container for the mobile. As an additional
// restriction it checks the Z distance against the uo.ContainerOpen* limits.
func openContainer(receiver, source, v any) bool {
	ri := receiver.(*game.Item)
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	rl := game.MapLocation(ri)
	sl := sm.Location
	dz := rl.Z - sl.Z
	if rl.XYDistance(sl) > uo.MaxUseRange {
		sm.NetState.Cliloc(nil, 500312) // You cannot reach that.
		return false
	}
	if dz < uo.ContainerOpenLowerLimit || dz > uo.ContainerOpenUpperLimit {
		sm.NetState.Cliloc(nil, 500312) // You cannot reach that.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(ri) {
		sm.NetState.Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	ri.Open(sm)
	return true
}

// mountMobile attempts to mount the source mobile onto the receiver.
func mountMobile(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if rm.ControlMaster != sm {
		return false
	}
	// Range check
	if sm.Location.XYDistance(rm.Location) > uo.MaxUseRange {
		sm.NetState.Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(rm) {
		sm.NetState.Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	sm.Mount(rm)
	return true
}

// openBackpack attempts to open the backpack of the receiver as in snooping or
// pack animals.
func openBackpack(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	rm := receiver.(*game.Mobile)
	// Control check
	if rm.ControlMaster != sm {
		// TODO Snooping?
		return false
	}
	// Range check
	if sm.Location.XYDistance(rm.Location) > uo.MaxUseRange {
		sm.NetState.Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(rm) {
		sm.NetState.Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	bp := rm.Equipment[uo.LayerBackpack]
	bp.Open(sm)
	return true
}

// openBankBox attempts to open the bank box of the source.
func openBankBox(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	bb := sm.Equipment[uo.LayerBankBox]
	bb.Open(sm)
	sm.NetState.Cliloc(receiver, 1080021, strconv.Itoa(bb.ItemCount),
		strconv.Itoa(int(bb.ContainedWeight))) // Bank container has ~1_VAL~ items, ~2_VAL~ stones
	return true
}

func transferHue(receiver, source, v any) bool {
	if receiver == nil || source == nil {
		return false
	}
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	var hue uo.Hue
	switch o := receiver.(type) {
	case *game.Mobile:
		hue = o.Hue
	case *game.Item:
		hue = o.Hue
	}
	sm.NetState.Speech(source, "Target object to set hue %d", hue)
	sm.NetState.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if m, found := game.World.FindMobile(tr.TargetObject); found {
			m.Hue = hue
			game.World.UpdateMobile(m)
		} else if i, found := game.World.FindItem(tr.TargetObject); found {
			i.Hue = hue
			game.World.UpdateItem(i)
		}
	})
	return true
}

func editObject(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil || !sm.Account.HasRole(game.RoleGameMaster) {
		return false
	}
	gumps.Edit(sm, receiver)
	return true
}

func openTeleportGUMP(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	sm.NetState.GUMP(gumps.New("teleport"), sm.Serial, 0)
	return true
}

func harvestCrop(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	ri := receiver.(*game.Item)
	// Range check
	if sm.Location.XYDistance(ri.Location) > uo.MaxUseRange {
		sm.NetState.Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(receiver) {
		sm.NetState.Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	tn := strings.TrimSuffix(ri.TemplateName, "Crop")
	i := game.NewItem(tn)
	if i == nil {
		return false
	}
	if !sm.DropToBackpack(i, false) {
		sm.DropToFeet(i)
	}
	game.World.RemoveItem(ri)
	return true
}

func cashCheck(receiver, source, v any) bool {
	check := receiver.(*game.Item)
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	if !sm.InBank(check) {
		sm.NetState.Speech(sm, "That must be in your bank box to use.")
		return false
	}
	game.World.RemoveItem(check)
	for {
		n := check.IArg
		if n < 1 {
			break
		}
		if n > uo.MaxStackAmount {
			n = uo.MaxStackAmount
		} else {
			check.IArg -= n
		}
		gc := game.NewItem("GoldCoin")
		gc.Amount = n
		sm.DropToBankBox(gc, true)
	}
	return true
}
