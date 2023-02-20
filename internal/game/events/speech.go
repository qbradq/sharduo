package events

import (
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("SpeechBanker", SpeechBanker)
}

// SpeechBanker handles banker speech triggers.
func SpeechBanker(receiver, source game.Object, v any) {
	if strings.Contains(strings.ToLower(v.(string)), "bank") {
		mob, ok := source.(game.Mobile)
		if !ok {
			return
		}
		bb := mob.EquipmentInSlot(uo.LayerBankBox)
		if bb == nil {
			return
		}
		box, ok := bb.(game.Container)
		if !ok {
			return
		}
		box.Open(mob)
	}
}
