package events

import "github.com/qbradq/sharduo/internal/game"

func init() {
	reg("ReportTime", ReportTime)
}

// ReportTime sends the current Sossarian time to the receiver as a system
// message. Mainly used in tests.
func ReportTime(receiver, source game.Object) {
	if m, ok := receiver.(game.Mobile); ok {
		if m.NetState() != nil {
			m.NetState().Speech(receiver, "%d", game.GetWorld().Time())
		}
	}
}
