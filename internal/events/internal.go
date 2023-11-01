package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("PlayerLogout", PlayerLogout)
	reg("WhisperTime", WhisperTime)
}

// WhisperTime whispers the current Sossarian time to the source.
func WhisperTime(receiver, source game.Object, v any) bool {
	if source == nil {
		return false
	}
	m, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if m.NetState() == nil {
		return false
	}
	m.NetState().Speech(source, "%d", game.GetWorld().Time())
	return true
}

// PlayerLogout logs out the player mobile receiver.
func PlayerLogout(receiver, source game.Object, v any) bool {
	if receiver == nil {
		return false
	}
	rm, ok := receiver.(game.Mobile)
	if !ok || rm.NetState() != nil {
		return false
	}
	game.GetWorld().Map().PlayEffect(uo.GFXTypeFixed, receiver, receiver, 0x3728,
		15, 10, true, false, uo.HueDefault, uo.GFXBlendModeNormal)
	game.GetWorld().Map().StoreObject(receiver)
	return true
}
