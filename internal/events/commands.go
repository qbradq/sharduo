package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Pet commands

func init() {
	reg("CommandDrop", commandDrop)
	reg("CommandFollow", commandFollow)
	reg("CommandFollowMe", commandFollowMe)
	reg("CommandRelease", commandRelease)
	reg("CommandStay", commandStay)
}

// commandFollow allows commanding a pet to follow another mobile with a
// targeting cursor.
func commandFollow(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	sm.NetState.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		tm, found := game.World.FindMobile(tr.TargetObject)
		if !found {
			return
		}
		rm.AI = "Follow"
		rm.AIGoal = tm
	})
	return true
}

// commandFollowMe commands a pet to follow the source mobile if that mobile can
// command the receiving mobile.
func commandFollowMe(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.AI = "Follow"
	rm.AIGoal = sm
	return true
}

// commandStay commands a pet to stay in its current location
// command the receiving mobile.
func commandStay(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.AI = "Stay"
	rm.AIGoal = nil
	return true
}

// commandDrop commands a pet to drop all inventory contents at their feet.
func commandDrop(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	bp := rm.Equipment[uo.LayerBackpack]
	if bp == nil {
		return false
	}
	items := make([]*game.Item, len(bp.Contents))
	copy(items, bp.Contents)
	for _, item := range items {
		bp.RemoveItem(item)
		rm.DropToFeet(item)
	}
	return true
}

// commandRelease commands a pet forget its control master.
func commandRelease(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.ControlMaster = nil
	rm.AI = "WalkRandom"
	rm.AIGoal = nil
	return true
}
