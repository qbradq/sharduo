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
func CommandFollow(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	sm.NetState().TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		tm := game.Find[game.Mobile](tr.TargetObject)
		if tm == nil {
			return
		}
		rm.SetAI("Follow")
		rm.SetAIGoal(tm)
	})
	return true
}

// CommandFollowMe commands a pet to follow the source mobile if that mobile can
// command the receiving mobile.
func CommandFollowMe(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.SetAI("Follow")
	rm.SetAIGoal(sm)
	return true
}

// CommandStay commands a pet to stay in its current location
// command the receiving mobile.
func CommandStay(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.SetAI("Stay")
	rm.SetAIGoal(nil)
	return true
}

// CommandDrop commands a pet to drop all inventory contents at their feet.
func CommandDrop(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	bpo := rm.EquipmentInSlot(uo.LayerBackpack)
	if bpo == nil {
		return false
	}
	bp, ok := bpo.(game.Container)
	if !ok {
		return false
	}
	contents := bp.Contents()
	items := make([]game.Item, len(contents))
	copy(items, contents)
	for _, item := range items {
		rm.DropToFeet(item)
	}
	return true
}

// CommandRelease commands a pet to drop all inventory contents at their feet.
func CommandRelease(receiver, source game.Object, v any) bool {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if !rm.CanBeCommandedBy(sm) {
		return false
	}
	rm.SetControlMaster(nil)
	rm.SetAI("WalkRandom")
	rm.SetAIGoal(nil)
	return true
}
