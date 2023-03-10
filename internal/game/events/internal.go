package events

import (
	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	reg("PlayerLogout", PlayerLogout)
	reg("WhisperTime", WhisperTime)
}

// WhisperTime whispers the current Sossarian time to the source.
func WhisperTime(receiver, source game.Object, v any) {
	if source == nil {
		return
	}
	m, ok := source.(game.Mobile)
	if !ok {
		return
	}
	if m.NetState() == nil {
		return
	}
	m.NetState().Speech(source, "%d", game.GetWorld().Time())
}

// PlayerLogout logs out the player mobile receiver.
func PlayerLogout(receiver, source game.Object, v any) {
	if receiver == nil {
		return
	}
	game.GetWorld().Map().RemoveObject(receiver)
}
