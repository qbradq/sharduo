package events

import "github.com/qbradq/sharduo/internal/game"

func init() {
	reg("WhisperTime", WhisperTime)
}

// WhisperTime whispers the current Sossarian time to the source.
func WhisperTime(receiver, source game.Object, v any) {
	m, ok := source.(game.Mobile)
	if !ok {
		return
	}
	if m.NetState() == nil {
		return
	}
	m.NetState().Speech(source, "%d", game.GetWorld().Time())
}
