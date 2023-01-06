package events

import (
	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	fnreg.Add("OpenPaperDoll", OpenPaperDoll)
}

// OpenPaperDoll opens the paper doll of the receiver mobile to the source.
func OpenPaperDoll(receiver, source game.Object) {
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	sm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	if sm.NetState() == nil {
		return
	}
	sm.NetState().OpenPaperDoll(rm)
}
