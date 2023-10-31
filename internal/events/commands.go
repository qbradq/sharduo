package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Pet commands

func init() {
	reg("CommandFollow", CommandFollow)
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
