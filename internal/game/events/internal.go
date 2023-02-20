package events

import "github.com/qbradq/sharduo/internal/game"

func init() {
	reg("WhisperTime", WhisperTime)
}

// WhisperTime whispers the current Sossarian time to the receiver.
func WhisperTime(receiver, source game.Object, v any) {
	m, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	if m.NetState() == nil {
		return
	}
	m.NetState().Speech(receiver, "%d", game.GetWorld().Time())
}
