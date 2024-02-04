package events

// Drop events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("DropToContainer", dropToContainer)
	reg("DropToPlayer", dropToPlayer)
	reg("DropToPackAnimal", dropToPackAnimal)
}

func dropToContainer(receiver, source, v any) bool {
	ri := receiver.(*game.Item)
	sm := source.(*game.Mobile)
	item := v.(*game.Item)
	if sm.Location.XYDistance(item.Location) > uo.MaxDropRange {
		return false
	}
	if !sm.CanAccess(ri) {
		return false
	}
	// Line of sight check
	if !sm.HasLineOfSight(ri) {
		sm.NetState.Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	if err := ri.AddItem(item, false); err != nil {
		e := err.(*game.UOError)
		sm.NetState.Send(e.Packet())
		return false
	}
	return true
}

func dropToPlayer(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	item := v.(*game.Item)
	if rm == sm {
		// Drop to self, just put it in the backpack
		return rm.DropToBackpack(item, false)
	} else if sm.Player {
		// TODO Secure trade
		return false
	} // Else this is a non-player mobile trying to drop something on us. This
	// should not happen.
	return false
}

func dropToPackAnimal(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	// Check control master
	if rm.ControlMaster != sm {
		return false
	}
	item := v.(*game.Item)
	return rm.DropToBackpack(item, false)
}
