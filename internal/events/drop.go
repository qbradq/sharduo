package events

// Drop events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("DropToContainer", DropToContainer)
	reg("DropToPlayer", DropToPlayer)
	reg("DropToPackAnimal", DropToPackAnimal)
}

func DropToContainer(receiver, source game.Object, v any) bool {
	rc, ok := receiver.(game.Container)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	item, ok := v.(game.Item)
	if !ok {
		return false
	}
	if sm.Location().XYDistance(item.Location()) > uo.MaxDropRange {
		return false
	}
	if !sm.CanAccess(rc) {
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(rc) {
		sm.NetState().Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	return rc.AddObject(item)
}

func DropToPlayer(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	item, ok := v.(game.Item)
	if !ok {
		return false
	}
	if rm.Serial() == sm.Serial() {
		// Drop to self, just put it in the backpack
		return rm.DropToBackpack(item, false)
	} else if sm.IsPlayerCharacter() {
		// TODO Secure trade
		return false
	} // Else this is a non-player mobile trying to drop something on us. This
	// should not happen.
	return false
}

func DropToPackAnimal(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	// Check control master
	if rm.ControlMaster().Serial() != sm.Serial() {
		return false
	}
	item, ok := v.(game.Item)
	if !ok {
		return false
	}
	return rm.DropToBackpack(item, false)
}
