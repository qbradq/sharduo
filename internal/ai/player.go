package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("Player", func() game.AIModel { return &player{} })
}

// player implements an AI that is responsible for periodic updates to player
// mobiles.
type player struct {
	lastMusicSent uo.Time
}

// Act implements the AIModel interface.
func (a *player) Act(m *game.Mobile, t uo.Time) {
	// Do nothing, ever
}

// Target implements the AIModel interface.
func (a *player) Target(m *game.Mobile, t uo.Time) {
	if m.NetState == nil {
		return
	}
	// Music triggers
	if t >= a.lastMusicSent+uo.DurationMinute*5 {
		exp := ""
		for _, r := range game.World.Map().RegionsAt(m.Location) {
			if len(r.Music) > 0 {
				exp = r.Music
			}
		}
		if len(exp) > 0 {
			m.NetState.Music(uo.Music(util.RangeExpression(exp)))
			a.lastMusicSent = t
		}
	}
}
