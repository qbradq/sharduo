package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Pet commands

func init() {
	reg("CommandDrop", CommandDrop)
	reg("CommandFollow", CommandFollow)
	reg("CommandFollowMe", CommandFollowMe)
	reg("CommandRelease", CommandRelease)
	reg("CommandStay", CommandStay)
}

// CommandFollow allows commanding a pet to follow another mobile with a
// targeting cursor.
func CommandFollow(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	sm.NetState.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		tm := game.World.FindMobile(tr.TargetObject)
		if tm == nil {
			return
		}
		rm.AI = "Follow"
		rm.AIGoal = tm
	})
	return true
}

// CommandFollowMe commands a pet to follow the source mobile if that mobile can
// command the receiving mobile.
func CommandFollowMe(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.AI = "Follow"
	rm.AIGoal = sm
	return true
}

// CommandStay commands a pet to stay in its current location
// command the receiving mobile.
func CommandStay(receiver, source, v any) bool {
	rm := receiver.(*game.Mobile)
	sm := source.(*game.Mobile)
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.AI = "Stay"
	rm.AIGoal = nil
	return true
}

// CommandDrop commands a pet to drop all inventory contents at their feet.
func CommandDrop(receiver, source, v any) bool {
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

// CommandRelease commands a pet forget its control master.
func CommandRelease(receiver, source, v any) bool {
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
