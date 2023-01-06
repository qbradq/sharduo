package events

// Common OnDoubleClick events

import (
	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	evreg.Add("OpenPaperDoll", OpenPaperDoll)
	evreg.Add("OpenContainer", OpenContainer)
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

// OpenContainer opens this container for the mobile.
func OpenContainer(receiver, source game.Object) {
	rc, ok := receiver.(game.Container)
	if !ok {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return
	}
	rc.Open(sm)
}
