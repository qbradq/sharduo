package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("PlayerLogout", playerLogout)
	reg("WhisperTime", whisperTime)
}

// whisperTime whispers the current Sossarian time to the source.
func whisperTime(receiver, source, v any) bool {
	sm := source.(game.Mobile)
	if sm.NetState == nil {
		return false
	}
	sm.NetState.Speech(source, "%d", game.World.Time())
	return true
}

// playerLogout logs out the player mobile receiver.
func playerLogout(receiver, source, v any) bool {
	rm := receiver.(game.Mobile)
	if rm.NetState != nil {
		return false
	}
	game.World.Map().PlayEffect(uo.GFXTypeFixed, receiver, receiver, 0x3728,
		15, 10, true, false, uo.HueDefault, uo.GFXBlendModeNormal)
	game.World.Map().StoreObject(receiver)
	return true
}
